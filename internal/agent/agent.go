package agent

import (
	"context"
	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
)

// RunAgent runs a single prompt against the configured model and returns
// the assistant output. The caller provides the systemPrompt string so this
// package does not need to import ui and create an import cycle.
func RunAgent(model string, prompt string, systemPrompt string, pol *types.Policy) (string, error) {
	msgs := []types.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}
	_, out, err := openai.ChatSession(context.Background(), model, msgs, pol)
	return out, err
}
