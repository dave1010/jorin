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
	"github.com/dave1010/jorin/internal/ralph"
	"github.com/dave1010/jorin/internal/repl"
	"github.com/dave1010/jorin/internal/repl/commands"
	"github.com/dave1010/jorin/internal/types"
)

// ErrMissingPrompt is returned when no prompt is provided and REPL is not requested.
var ErrMissingPrompt = errors.New("provide a prompt or use --repl")

// Config configures the application run.
type Config struct {
	Model           string
	Prompt          string
	Repl            bool
	NoArgs          bool
	ScriptArgs      []string
	RalphMaxTries   int
	Policy          types.Policy
	Stdin           io.Reader
	StdinIsTTY      bool
	Stdout          io.Writer
	Stderr          io.Writer
	UseResponsesAPI bool
}

// App holds the application's dependencies.
type App struct {
	cfg     *Config
	agent   agent.Agent
	history repl.History
}

// NewApp creates a new App with the given configuration.
func NewApp(cfg *Config) *App {
	plugins.SetModelProvider(func() string { return cfg.Model })

	return &App{
		cfg:     cfg,
		agent:   openai.NewDefaultAgent(cfg.UseResponsesAPI),
		history: repl.NewMemHistory(200),
	}
}

// Run wires core dependencies and starts either the REPL or a single prompt run.
func (a *App) Run(ctx context.Context) error {
	if a.cfg.NoArgs || a.cfg.Repl {
		return a.runRepl(ctx)
	}
	return a.runPrompt()
}

func (a *App) runRepl(ctx context.Context) error {
	cfg := repl.DefaultConfig()
	handler := commands.NewDefaultHandler(a.cfg.Stdout, a.cfg.Stderr, a.history, prompt.SystemPrompt)

	return repl.StartREPL(repl.StartOptions{
		Ctx:     ctx,
		Agent:   a.agent,
		Model:   a.cfg.Model,
		Policy:  &a.cfg.Policy,
		Input:   a.cfg.Stdin,
		Output:  a.cfg.Stdout,
		ErrOut:  a.cfg.Stderr,
		Config:  cfg,
		Handler: handler,
		History: a.history,
	})
}

func (a *App) runPrompt() error {
	stdinText, err := readPromptStdin(a.cfg)
	if err != nil {
		return err
	}
	fullPrompt := buildPrompt(a.cfg.Prompt, a.cfg.ScriptArgs, stdinText)
	if strings.TrimSpace(fullPrompt) == "" {
		return ErrMissingPrompt
	}

	systemPrompt := prompt.SystemPrompt()
	if prompt.RalphEnabled() {
		if err := ralph.Run(a.agent, a.cfg.Model, fullPrompt, systemPrompt, &a.cfg.Policy, a.cfg.RalphMaxTries, a.cfg.Stdout, a.cfg.Stderr); err != nil {
			return err
		}
		return nil
	}

	msgs := []types.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fullPrompt},
	}
	_, out, err := a.agent.ChatSession(a.cfg.Model, msgs, &a.cfg.Policy)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(a.cfg.Stdout, out); err != nil {
		return err
	}
	return nil
}

func readPromptStdin(cfg *Config) (string, error) {
	if cfg.StdinIsTTY || cfg.Stdin == nil {
		return "", nil
	}
	data, err := io.ReadAll(cfg.Stdin)
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
