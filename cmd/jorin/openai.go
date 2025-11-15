package main

import (
	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
)

// Wrapper to re-export ChatSession from internal/openai for the main package.
// This keeps main package light while delegating implementation to internal.
// Keep this function referenced by main to avoid unused lint when file exists.
func chatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return openai.ChatSession(model, msgs, pol)
}
