package main

import (
	"github.com/dave1010/jorin/internal/openai"
)

func chatOnce(model string, msgs []Message, tools []Tool) (*ChatResponse, error) {
	return openai.ChatOnce(model, msgs, tools)
}

func chatSession(model string, msgs []Message, pol *Policy) ([]Message, string, error) {
	return openai.ChatSession(model, msgs, pol)
}
