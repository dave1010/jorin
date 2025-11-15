package main

import (
	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
)

// Wrapper functions to re-export functionality from internal/openai for the
// main package. These are kept for a simple, small public API surface in main
// while delegating actual logic to the internal package.
func chatOnce(model string, msgs []types.Message, tools []types.Tool) (*types.ChatResponse, error) {
	return openai.ChatOnce(model, msgs, tools)
}

func chatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return openai.ChatSession(model, msgs, pol)
}
