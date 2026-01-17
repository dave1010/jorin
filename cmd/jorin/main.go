package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dave1010/jorin/internal/app"
	"github.com/dave1010/jorin/internal/types"
	"github.com/dave1010/jorin/internal/version"
)

// reference chatSession to avoid unused lint when file exists
var _ = chatSession

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

	prompt := stringJoin(flag.Args(), " ")
	opts := app.Options{
		Model:  *model,
		Prompt: prompt,
		Repl:   *repl,
		NoArgs: len(os.Args) == 1,
		Policy: types.Policy{
			Readonly: *readonly,
			DryShell: *dry,
			Allow:    *allow,
			Deny:     *deny,
			CWD:      *cwd,
		},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := app.Run(context.Background(), opts); err != nil {
		if err == app.ErrMissingPrompt {
			fmt.Fprintln(os.Stderr, "Provide a prompt or use --repl")
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
}

// wrappers so main doesn't import strings package etc.
func stringJoin(a []string, sep string) string { return strings.Join(a, sep) }

// keep multi flag helpers
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }
func multi(name, usage string) *multiFlag {
	var v multiFlag
	flag.Var(&v, name, usage)
	return &v
}
