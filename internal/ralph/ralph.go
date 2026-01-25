package ralph

import (
	"fmt"
	"io"
	"strings"

	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/types"
)

// Run runs the Ralph Wiggum loop.
func Run(ag agent.Agent, model string, initialPrompt string, systemPrompt string, pol *types.Policy, maxTries int, stdout io.Writer, stderr io.Writer) error {
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
		msgs := []types.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: currentPrompt},
		}
		_, out, err := ag.ChatSession(model, msgs, pol)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(stdout, out); err != nil {
			return err
		}
		if Done(out) {
			return nil
		}
		currentPrompt = out
	}
	return fmt.Errorf("ralph loop reached max tries (%d) without DONE", maxTries)
}

// Done checks if the output of a Ralph Wiggum loop indicates that it is done.
func Done(output string) bool {
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
