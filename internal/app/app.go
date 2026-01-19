package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/plugins"
	"github.com/dave1010/jorin/internal/prompt"
	"github.com/dave1010/jorin/internal/repl"
	"github.com/dave1010/jorin/internal/repl/commands"
	"github.com/dave1010/jorin/internal/types"
)

// ErrMissingPrompt is returned when no prompt is provided and REPL is not requested.
var ErrMissingPrompt = errors.New("provide a prompt or use --repl")

// Options configures the application run.
type Options struct {
	Model      string
	Prompt     string
	Repl       bool
	NoArgs     bool
	ScriptArgs []string
	Policy     types.Policy
	Stdin      io.Reader
	StdinIsTTY bool
	Stdout     io.Writer
	Stderr     io.Writer
}

// Run wires core dependencies and starts either the REPL or a single prompt run.
func Run(ctx context.Context, opts Options) error {
	plugins.SetModelProvider(func() string { return opts.Model })

	cfg := repl.DefaultConfig()
	hist := repl.NewMemHistory(200)
	handler := commands.NewDefaultHandler(opts.Stdout, opts.Stderr, hist, prompt.SystemPrompt)

	var agentImpl agent.Agent = &openai.DefaultAgent{}

	if opts.NoArgs || opts.Repl {
		return repl.StartREPL(ctx, agentImpl, opts.Model, &opts.Policy, opts.Stdin, opts.Stdout, opts.Stderr, cfg, handler, hist)
	}

	stdinText := ""
	if !opts.StdinIsTTY && opts.Stdin != nil {
		data, err := io.ReadAll(opts.Stdin)
		if err != nil {
			return err
		}
		stdinText = string(data)
	}

	fullPrompt := buildPrompt(opts.Prompt, opts.ScriptArgs, stdinText)
	if strings.TrimSpace(fullPrompt) == "" {
		return ErrMissingPrompt
	}

	out, err := agent.RunAgent(opts.Model, fullPrompt, prompt.SystemPrompt(), &opts.Policy)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(opts.Stdout, out); err != nil {
		return err
	}
	return nil
}

func buildPrompt(prompt string, args []string, stdin string) string {
	parts := []string{}
	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt != "" {
		parts = append(parts, prompt)
	}
	if len(args) > 0 {
		parts = append(parts, "Arguments: "+strings.Join(args, " "))
	}
	if trimmedPrompt == "" && len(args) == 0 && stdin != "" {
		return strings.TrimRight(stdin, "\n")
	}
	if stdin != "" {
		parts = append(parts, "Stdin:\n"+strings.TrimRight(stdin, "\n"))
	}
	return strings.Join(parts, "\n\n")
}
