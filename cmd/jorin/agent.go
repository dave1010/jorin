package main

import (
	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/types"
)

// runAgent delegates to the internal/agent package.
func runAgent(model string, userPrompt string, pol *types.Policy) (string, error) {
	return agent.RunAgent(model, userPrompt, pol)
}
