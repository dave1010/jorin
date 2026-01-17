package repl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/prompt"
	"github.com/dave1010/jorin/internal/repl/commands"
	"github.com/dave1010/jorin/internal/tools"
	"github.com/dave1010/jorin/internal/types"
)

// StartREPL runs an interactive REPL using the provided reader/writer. It is
// testable because IO is injected. It accepts a commands.Handler and a History
// implementation so command dispatch and history persistence are pluggable.
func StartREPL(ctx context.Context, a agent.Agent, model string, pol *types.Policy, in io.Reader, out io.Writer, errOut io.Writer, cfg *Config, handler commands.Handler, hist History) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if _, err := fmt.Fprintln(out, headerStyleStr("jorin\u003e (Ctrl-D to exit)")); err != nil {
		return err
	}
	msgs := []types.Message{{Role: "system", Content: prompt.SystemPrompt()}}
	reg := tools.Registry()

	// create a LineReader that provides proper terminal editing when possible
	lr := NewLineReader(in, out)
	defer func() { _ = lr.Close() }()
	if hist != nil {
		// append previous history so arrow-up works for past sessions
		lr.AppendHistory(hist.List(0))
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line, err := lr.ReadLine(promptStyleStr(cfg.Prompt))
		if err != nil {
			if err == io.EOF {
				break
			}
			if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), err); werr != nil {
				return werr
			}
			continue
		}
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		// attempt to parse as slash command
		cmd, perr := commands.Parse(trim, cfg.CommandPrefix, cfg.EscapePrefix)
		if perr == nil {
			// parsed a command; dispatch to handler
			handled, err := handler.Handle(ctx, cmd)
			if err != nil {
				if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), err); werr != nil {
					return werr
				}
				continue
			}
			if handled {
				// do not forward to model
				continue
			}
			// if not handled, fallthrough and forward command text
			trim = cmd.Raw
		} else {
			// if escaped, Parse returns an error but sets Raw to the unescaped
			// literal. The parser currently uses an "escaped" error string. If
			// the parser indicated escaped, use the Raw content. Otherwise it's
			// not a command and we forward the original trimmed line.
			if perr.Error() == "escaped" {
				trim = cmd.Raw
			}
		}
		// legacy: support leading '!' shell commands via tools registry
		if strings.HasPrefix(trim, "!") {
			cmdStr := strings.TrimSpace(trim[1:])
			if cmdStr == "" {
				if _, werr := fmt.Fprintln(errOut, infoStyleStr("empty shell command")); werr != nil {
					return werr
				}
				continue
			}
			if sh, ok := reg["shell"]; ok {
				res, err := sh(map[string]any{"cmd": cmdStr}, pol)
				if err != nil {
					if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), err); werr != nil {
						return werr
					}
					continue
				}
				if e, ok := res["error"]; ok {
					if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), e); werr != nil {
						return werr
					}
					continue
				}
				if dr, _ := res["dry_run"].(bool); dr {
					if c, ok := res["cmd"].(string); ok {
						if _, werr := fmt.Fprintln(errOut, infoStyleStr("Dry run:"), c); werr != nil {
							return werr
						}
					} else {
						if _, werr := fmt.Fprintln(errOut, infoStyleStr("Dry run:"), cmdStr); werr != nil {
							return werr
						}
					}
					continue
				}
				if sout, ok := res["stdout"].(string); ok && sout != "" {
					if _, werr := fmt.Fprintln(out, infoStyleStr(sout)); werr != nil {
						return werr
					}
				}
				if serr, ok := res["stderr"].(string); ok && serr != "" {
					if _, werr := fmt.Fprintln(errOut, errorStyleStr(serr)); werr != nil {
						return werr
					}
				}
				if rc, ok := res["returncode"]; ok {
					if _, werr := fmt.Fprintln(errOut, infoStyleStr("returncode:"), rc); werr != nil {
						return werr
					}
				}
				continue
			}
			if _, werr := fmt.Fprintln(errOut, errorStyleStr("shell tool not available")); werr != nil {
				return werr
			}
			continue
		}
		// forward to model via agent interface
		msgs = append(msgs, types.Message{Role: "user", Content: trim})
		if hist != nil {
			hist.Add(trim)
		}
		var outStr string
		var err2 error
		msgs, outStr, err2 = a.ChatSession(model, msgs, pol)
		if err2 != nil {
			if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), err2); werr != nil {
				return werr
			}
			continue
		}
		if _, werr := fmt.Fprintln(out, infoStyleStr(outStr)); werr != nil {
			return werr
		}
	}
	return nil
}

// The following helper style functions are duplicated from cmd to avoid a
// dependency cycle; they intentionally return plain strings so callers can
// decide where to write them.
func headerStyleStr(s string) string { return s }
func promptStyleStr(s string) string { return s }
func infoStyleStr(s string) string   { return s }
func errorStyleStr(s string) string  { return s }
