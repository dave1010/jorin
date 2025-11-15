package main

import (
	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/types"
	"github.com/dave1010/jorin/internal/ui"
)

// runAgent delegates to the internal/agent package and passes system prompt.
func runAgent(model string, userPrompt string, pol *types.Policy) (string, error) {
	return agent.RunAgent(model, userPrompt, ui.SystemPrompt(), pol)
}
