package main

import (
	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
)

func chatOnce(model string, msgs []types.Message, tools []types.Tool) (*types.ChatResponse, error) {
	return openai.ChatOnce(model, msgs, tools)
}

func chatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return openai.ChatSession(model, msgs, pol)
}
