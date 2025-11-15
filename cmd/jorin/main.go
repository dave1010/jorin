package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
	"github.com/dave1010/jorin/internal/ui"
	cmdcommands "github.com/dave1010/jorin/internal/ui/commands"
	"github.com/dave1010/jorin/internal/version"
)

func main() {
	model := flag.String("model", "gpt-5-mini", "Model ID")
	repl := flag.Bool("repl", false, "Interactive REPL")
	readonly := flag.Bool("readonly", false, "Disallow write_file")
	dry := flag.Bool("dry-shell", false, "Do not execute shell commands")
	allow := multi("allow", "Allowlist substring for shell (repeatable)")
	deny := multi("deny", "Denylist substring for shell (repeatable)")
	cwd := flag.String("cwd", "", "Working directory for tools")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	pol := &types.Policy{Readonly: *readonly, DryShell: *dry, Allow: *allow, Deny: *deny, CWD: *cwd}

	// prepare handler and history
	cfg := ui.DefaultConfig()
	hist := ui.NewMemHistory(200)
	handler := cmdcommands.NewDefaultHandler(os.Stdout, os.Stderr, hist)
	// default agent based on openai package
	agentImpl := &openai.DefaultAgent{}

	// If program invoked with no args at all, behave as if --repl was provided.
	if len(os.Args) == 1 {
		ui.StartREPL(context.Background(), agentImpl, *model, pol, os.Stdin, os.Stdout, os.Stderr, cfg, handler, hist)
		return
	}

	if *repl {
		ui.StartREPL(context.Background(), agentImpl, *model, pol, os.Stdin, os.Stdout, os.Stderr, cfg, handler, hist)
		return
	}

	prompt := stringJoin(flag.Args(), " ")
	if stringTrimSpace(prompt) == "" {
		fmt.Fprintln(os.Stderr, "Provide a prompt or use --repl")
		os.Exit(2)
	}
	// fallback single-run prompt uses agent.RunAgent convenience; pass system prompt
	out, err := agent.RunAgent(*model, prompt, ui.SystemPrompt(), pol)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

// wrappers so main doesn't import strings package etc.
func stringJoin(a []string, sep string) string { return strings.Join(a, sep) }
func stringTrimSpace(s string) string          { return strings.TrimSpace(s) }

// keep multi flag helpers
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }
func multi(name, usage string) *multiFlag {
	var v multiFlag
	flag.Var(&v, name, usage)
	return &v
}
