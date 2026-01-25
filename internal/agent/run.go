package agent

import (
	"github.com/dave1010/jorin/internal/prompt"
	"github.com/dave1010/jorin/internal/types"
)

// RunWithSystemPrompt delegates to the internal/agent package and passes system prompt.
func RunWithSystemPrompt(model string, userPrompt string, pol *types.Policy) (string, error) {
	return RunAgent(model, userPrompt, prompt.SystemPrompt(), pol)
}
