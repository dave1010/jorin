package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dave1010/jorin/internal/app"
	"github.com/dave1010/jorin/internal/prompt"
	"github.com/dave1010/jorin/internal/types"
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
	promptFlag := flag.Bool("prompt", false, "Treat first argument as prompt text")
	promptFileFlag := flag.Bool("prompt-file", false, "Treat first argument as a prompt file")
	ralph := flag.Bool("ralph", false, "Enable Ralph Wiggum loop instructions")
	ralphMaxTries := flag.Int("ralph-max-tries", 8, "Maximum Ralph Wiggum loop iterations")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	if *promptFlag && *promptFileFlag {
		fmt.Fprintln(os.Stderr, "ERR: --prompt and --prompt-file cannot be used together")
		os.Exit(2)
	}

	if *ralph {
		prompt.EnableRalph()
	}
	if *ralphMaxTries < 1 {
		fmt.Fprintln(os.Stderr, "ERR: --ralph-max-tries must be at least 1")
		os.Exit(2)
	}

	promptMode := promptModeAuto
	if *promptFlag {
		promptMode = promptModeText
	} else if *promptFileFlag {
		promptMode = promptModeFile
	}

	stdinIsTTY := isTTY(os.Stdin)
	prompt, scriptArgs, err := resolvePrompt(flag.Args(), promptMode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
	noArgs := len(flag.Args()) == 0 && stdinIsTTY

	opts := app.Options{
		Model:         *model,
		Prompt:        prompt,
		Repl:          *repl,
		NoArgs:        noArgs,
		ScriptArgs:    scriptArgs,
		RalphMaxTries: *ralphMaxTries,
		Policy: types.Policy{
			Readonly: *readonly,
			DryShell: *dry,
			Allow:    *allow,
			Deny:     *deny,
			CWD:      *cwd,
		},
		Stdin:      os.Stdin,
		StdinIsTTY: stdinIsTTY,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
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

// keep multi flag helpers
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }
func multi(name, usage string) *multiFlag {
	var v multiFlag
	flag.Var(&v, name, usage)
	return &v
}
