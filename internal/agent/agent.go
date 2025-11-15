package agent

import (
	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
	"github.com/dave1010/jorin/internal/ui"
)

// RunAgent runs a single prompt against the configured model and returns
// the assistant output. It composes the system prompt from the UI package
// (which includes runtime/project context) and delegates session handling
// to the openai package.
func RunAgent(model string, prompt string, pol *types.Policy) (string, error) {
	msgs := []types.Message{
		{Role: "system", Content: ui.SystemPrompt()},
		{Role: "user", Content: prompt},
	}
	_, out, err := openai.ChatSession(model, msgs, pol)
	return out, err
}
