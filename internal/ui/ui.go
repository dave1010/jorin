package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/tools"
	"github.com/dave1010/jorin/internal/types"
)

const systemPromptBase = `You are a coding agent, designed to call tools to complete tasks.
Respond either with a normal assistant message, or with tool calls (function calling).
Prefer small, auditable steps. Read before you write. Don't suggest extra work.

## Git

Only run git commands if explicitly asked.

'git add .' is verboten. Always add paths intentionally.
`

// SystemPrompt builds the base system prompt and appends project-specific
// instructions from AGENTS.md if present, plus runtime context information.
func SystemPrompt() string {
	sp := systemPromptBase
	if _, err := os.Stat("AGENTS.md"); err == nil {
		if b, err := os.ReadFile("AGENTS.md"); err == nil {
			sp = sp + "\n\nProject-specific instructions:\n" + string(b)
		}
	}
	if ctx := runtimeContext(); ctx != "" {
		sp = sp + "\n\nRuntime environment:\n" + ctx
	}
	return sp
}

func runtimeContext() string {
	parts := []string{}
	if out, err := exec.Command("uname", "-a").Output(); err == nil {
		parts = append(parts, strings.TrimSpace(string(out)))
	} else {
		parts = append(parts, "OS: "+runtime.GOOS+" "+runtime.GOARCH)
	}
	if wd, err := os.Getwd(); err == nil {
		parts = append(parts, "PWD: "+wd)
	}
	if _, err := os.Stat(".git"); err == nil {
		parts = append(parts, "Git repository: yes (.git exists)")
	} else {
		parts = append(parts, "Git repository: no (.git not found)")
	}
	toolsList := []string{"ag", "rg", "git", "gh", "go", "gofmt", "docker", "fzf", "python", "python3", "php", "curl", "wget"}
	found := []string{}
	for _, t := range toolsList {
		if _, err := exec.LookPath(t); err == nil {
			found = append(found, t+" ")
		}
	}
	if len(found) > 0 {
		parts = append(parts, "Tools on PATH (others will exist too): "+strings.Join(found, ", "))
	} else {
		parts = append(parts, "Tools on PATH: none of "+strings.Join(toolsList, ", "))
	}
	return strings.Join(parts, "\n")
}

// StartREPL runs an interactive REPL using the provided reader/writer. It is
// testable because IO is injected. It uses the openai.ChatSession to talk to
// the model and the tools registry from internal/tools for local actions.
func StartREPL(model string, pol *types.Policy, in io.Reader, out io.Writer, errOut io.Writer) {
	scanner := bufio.NewScanner(in)
	fmt.Fprintln(out, headerStyleStr("jorin> (Ctrl-D to exit)"))
	msgs := []types.Message{{Role: "system", Content: SystemPrompt()}}
	reg := tools.Registry()
	for {
		fmt.Fprint(out, promptStyleStr("> "))
		if !scanner.Scan() {
			break
		}
		q := strings.TrimSpace(scanner.Text())
		if q == "" {
			continue
		}
		if strings.HasPrefix(q, "/") {
			if strings.HasPrefix(q, "/debug") {
				sp := SystemPrompt()
				fmt.Fprintln(errOut, infoStyleStr(sp))
				continue
			}
			fmt.Fprintln(errOut, infoStyleStr("unknown command:"), q)
			continue
		}
		if strings.HasPrefix(q, "!") {
			cmdStr := strings.TrimSpace(q[1:])
			if cmdStr == "" {
				fmt.Fprintln(errOut, infoStyleStr("empty shell command"))
				continue
			}
			if sh, ok := reg["shell"]; ok {
				res, err := sh(map[string]any{"cmd": cmdStr}, pol)
				if err != nil {
					fmt.Fprintln(errOut, errorStyleStr("ERR:"), err)
					continue
				}
				if e, ok := res["error"]; ok {
					fmt.Fprintln(errOut, errorStyleStr("ERR:"), e)
					continue
				}
				if dr, ok := res["dry_run"].(bool); ok && dr {
					if c, ok := res["cmd"].(string); ok {
						fmt.Fprintln(errOut, infoStyleStr("Dry run:"), c)
					} else {
						fmt.Fprintln(errOut, infoStyleStr("Dry run:"), cmdStr)
					}
					continue
				}
				if sout, ok := res["stdout"].(string); ok && sout != "" {
					fmt.Fprintln(out, infoStyleStr(sout))
				}
				if serr, ok := res["stderr"].(string); ok && serr != "" {
					fmt.Fprintln(errOut, errorStyleStr(serr))
				}
				if rc, ok := res["returncode"]; ok {
					fmt.Fprintln(errOut, infoStyleStr("returncode:"), rc)
				}
				continue
			}
			fmt.Fprintln(errOut, errorStyleStr("shell tool not available"))
			continue
		}
		msgs = append(msgs, types.Message{Role: "user", Content: q})
		var outStr string
		var err error
		msgs, outStr, err = openai.ChatSession(model, msgs, pol)
		if err != nil {
			fmt.Fprintln(errOut, errorStyleStr("ERR:"), err)
			continue
		}
		fmt.Fprintln(out, infoStyleStr(outStr))
	}
}

// The following helper style functions are duplicated from cmd to avoid a
// dependency cycle; they intentionally return plain strings so callers can
// decide where to write them.
func headerStyleStr(s string) string { return s }
func promptStyleStr(s string) string { return s }
func infoStyleStr(s string) string   { return s }
func errorStyleStr(s string) string  { return s }
