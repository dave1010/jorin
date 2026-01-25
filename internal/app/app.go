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
	Model         string
	Prompt        string
	Repl          bool
	NoArgs        bool
	ScriptArgs    []string
	RalphMaxTries int
	Policy        types.Policy
	Stdin         io.Reader
	StdinIsTTY    bool
	Stdout        io.Writer
	Stderr        io.Writer
}

// Run wires core dependencies and starts either the REPL or a single prompt run.
func Run(ctx context.Context, opts Options) error {
	plugins.SetModelProvider(func() string { return opts.Model })

	cfg := repl.DefaultConfig()
	hist := repl.NewMemHistory(200)
	handler := commands.NewDefaultHandler(opts.Stdout, opts.Stderr, hist, prompt.SystemPrompt)

	var agentImpl agent.Agent = &openai.DefaultAgent{}

	if opts.NoArgs || opts.Repl {
		return runRepl(ctx, opts, cfg, handler, hist, agentImpl)
	}
	return runPrompt(opts)
}

func runRepl(ctx context.Context, opts Options, cfg *repl.Config, handler commands.Handler, hist repl.History, agentImpl agent.Agent) error {
	return repl.StartREPL(repl.StartOptions{
		Ctx:     ctx,
		Agent:   agentImpl,
		Model:   opts.Model,
		Policy:  &opts.Policy,
		Input:   opts.Stdin,
		Output:  opts.Stdout,
		ErrOut:  opts.Stderr,
		Config:  cfg,
		Handler: handler,
		History: hist,
	})
}

func runPrompt(opts Options) error {
	stdinText, err := readPromptStdin(opts)
	if err != nil {
		return err
	}
	fullPrompt := buildPrompt(opts.Prompt, opts.ScriptArgs, stdinText)
	if strings.TrimSpace(fullPrompt) == "" {
		return ErrMissingPrompt
	}

	systemPrompt := prompt.SystemPrompt()
	if prompt.RalphEnabled() {
		if err := runRalphLoop(opts.Model, fullPrompt, systemPrompt, &opts.Policy, opts.RalphMaxTries, opts.Stdout, opts.Stderr); err != nil {
			return err
		}
		return nil
	}

	out, err := agent.RunAgent(opts.Model, fullPrompt, systemPrompt, &opts.Policy)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(opts.Stdout, out); err != nil {
		return err
	}
	return nil
}

func readPromptStdin(opts Options) (string, error) {
	if opts.StdinIsTTY || opts.Stdin == nil {
		return "", nil
	}
	data, err := io.ReadAll(opts.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
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

func runRalphLoop(model string, initialPrompt string, systemPrompt string, pol *types.Policy, maxTries int, stdout io.Writer, stderr io.Writer) error {
	if maxTries < 1 {
		return fmt.Errorf("ralph max tries must be at least 1")
	}
	currentPrompt := initialPrompt
	for i := 0; i < maxTries; i++ {
		if stderr != nil {
			if _, err := fmt.Fprintf(stderr, "Ralph iteration %d/%d\n", i+1, maxTries); err != nil {
				return err
			}
		}
		out, err := agent.RunAgent(model, currentPrompt, systemPrompt, pol)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(stdout, out); err != nil {
			return err
		}
		if ralphDone(out) {
			return nil
		}
		currentPrompt = out
	}
	return fmt.Errorf("ralph loop reached max tries (%d) without DONE", maxTries)
}

func ralphDone(output string) bool {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		return line == "DONE"
	}
	return false
}
