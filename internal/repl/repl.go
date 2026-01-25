package repl

import (
	"context"
	"errors"
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
type StartOptions struct {
	Ctx     context.Context
	Agent   agent.Agent
	Model   string
	Policy  *types.Policy
	Input   io.Reader
	Output  io.Writer
	ErrOut  io.Writer
	Config  *Config
	Handler commands.Handler
	History History
}

func StartREPL(opts StartOptions) error {
	if opts.Config == nil {
		opts.Config = DefaultConfig()
	}
	if _, err := fmt.Fprintln(opts.Output, headerStyleStr("jorin\u003e (Ctrl-D to exit)")); err != nil {
		return err
	}
	msgs := []types.Message{{Role: "system", Content: prompt.SystemPrompt()}}
	reg := tools.Registry()

	// create a LineReader that provides proper terminal editing when possible
	lr := NewLineReader(opts.Input, opts.Output)
	defer func() { _ = lr.Close() }()
	if opts.History != nil {
		// append previous history so arrow-up works for past sessions
		lr.AppendHistory(opts.History.List(0))
	}

	for {
		select {
		case <-opts.Ctx.Done():
			return opts.Ctx.Err()
		default:
		}
		line, done, err := readPromptLine(lr, promptStyleStr(opts.Config.Prompt))
		if err != nil {
			if _, werr := fmt.Fprintln(opts.ErrOut, errorStyleStr("ERR:"), err); werr != nil {
				return werr
			}
			continue
		}
		if done {
			break
		}
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		trim, handled, err := parseAndHandleCommand(opts.Ctx, trim, opts.Config, opts.Handler)
		if err != nil {
			if _, werr := fmt.Fprintln(opts.ErrOut, errorStyleStr("ERR:"), err); werr != nil {
				return werr
			}
			continue
		}
		if handled {
			continue
		}
		handled, err = handleShellCommand(trim, reg, opts.Policy, opts.Output, opts.ErrOut)
		if err != nil {
			return err
		}
		if handled {
			continue
		}
		msgs, err = forwardToAgent(opts.Agent, opts.Model, trim, opts.Policy, opts.History, msgs, opts.Output, opts.ErrOut)
		if err != nil {
			return err
		}
	}
	return nil
}

func readPromptLine(lr LineReader, prompt string) (string, bool, error) {
	line, err := lr.ReadLine(prompt)
	if err != nil {
		if err == io.EOF {
			return "", true, nil
		}
		return "", false, err
	}
	return line, false, nil
}

func parseAndHandleCommand(ctx context.Context, line string, cfg *Config, handler commands.Handler) (string, bool, error) {
	cmd, err := commands.Parse(line, cfg.CommandPrefix, cfg.EscapePrefix)
	if err == nil {
		handled, handleErr := handler.Handle(ctx, cmd)
		if handleErr != nil {
			return "", true, handleErr
		}
		if handled {
			return "", true, nil
		}
		return cmd.Raw, false, nil
	}
	if errors.Is(err, commands.ErrEscaped) {
		return cmd.Raw, false, nil
	}
	return line, false, nil
}

func handleShellCommand(line string, reg map[string]tools.ToolExec, pol *types.Policy, out io.Writer, errOut io.Writer) (bool, error) {
	if !strings.HasPrefix(line, "!") {
		return false, nil
	}
	cmdStr := strings.TrimSpace(line[1:])
	if cmdStr == "" {
		if _, werr := fmt.Fprintln(errOut, infoStyleStr("empty shell command")); werr != nil {
			return true, werr
		}
		return true, nil
	}
	sh, ok := reg["shell"]
	if !ok {
		if _, werr := fmt.Fprintln(errOut, errorStyleStr("shell tool not available")); werr != nil {
			return true, werr
		}
		return true, nil
	}
	res, err := sh(map[string]any{"cmd": cmdStr}, pol)
	if err != nil {
		if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), err); werr != nil {
			return true, werr
		}
		return true, nil
	}
	if err := reportShellResult(cmdStr, res, out, errOut); err != nil {
		return true, err
	}
	return true, nil
}

func reportShellResult(cmdStr string, res map[string]any, out io.Writer, errOut io.Writer) error {
	if e, ok := res["error"]; ok {
		if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), e); werr != nil {
			return werr
		}
		return nil
	}
	if dr, _ := res["dry_run"].(bool); dr {
		cmd := cmdStr
		if c, ok := res["cmd"].(string); ok && c != "" {
			cmd = c
		}
		if _, werr := fmt.Fprintln(errOut, infoStyleStr("Dry run:"), cmd); werr != nil {
			return werr
		}
		return nil
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
	return nil
}

func forwardToAgent(a agent.Agent, model string, line string, pol *types.Policy, hist History, msgs []types.Message, out io.Writer, errOut io.Writer) ([]types.Message, error) {
	msgs = append(msgs, types.Message{Role: "user", Content: line})
	if hist != nil {
		hist.Add(line)
	}
	var outStr string
	var err error
	msgs, outStr, err = a.ChatSession(model, msgs, pol)
	if err != nil {
		if _, werr := fmt.Fprintln(errOut, errorStyleStr("ERR:"), err); werr != nil {
			return msgs, werr
		}
		return msgs, nil
	}
	if _, werr := fmt.Fprintln(out, infoStyleStr(outStr)); werr != nil {
		return msgs, werr
	}
	return msgs, nil
}

// The following helper style functions are duplicated from cmd to avoid a
// dependency cycle; they intentionally return plain strings so callers can
// decide where to write them.
func headerStyleStr(s string) string { return s }
func promptStyleStr(s string) string { return s }
func infoStyleStr(s string) string   { return s }
func errorStyleStr(s string) string  { return s }
