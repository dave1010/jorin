package main

import (
	"context"
	"fmt"
	flag "github.com/spf13/pflag"
	"os"

	"github.com/dave1010/jorin/internal/app"
	"github.com/dave1010/jorin/internal/types"
)

func main() {
	cli := parseFlags()
	handlePreflight(cli)

	promptMode := resolvePromptMode(cli.promptFlag, cli.promptFileFlag)
	stdinIsTTY := isTTY(os.Stdin)
	promptText, scriptArgs, err := resolvePrompt(flag.Args(), promptMode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
	noArgs := len(flag.Args()) == 0 && stdinIsTTY

	cfg := app.Config{
		Model:           cli.model,
		Prompt:          promptText,
		Repl:            cli.repl,
		NoArgs:          noArgs,
		ScriptArgs:      scriptArgs,
		RalphMaxTries:   cli.ralphMaxTries,
		UseResponsesAPI: cli.useResponsesAPI,
		Policy: types.Policy{
			Readonly: cli.readonly,
			DryShell: cli.dryShell,
			Allow:    cli.allow,
			Deny:     cli.deny,
			CWD:      cli.cwd,
		},
		Stdin:      os.Stdin,
		StdinIsTTY: stdinIsTTY,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}
	if err := app.NewApp(&cfg).Run(context.Background()); err != nil {
		exitWithError(err)
	}
}
