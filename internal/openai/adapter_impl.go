package openai

import "github.com/dave1010/jorin/internal/types"

// DefaultAgent implements types.Agent by delegating to package-level
// ChatSession.
type DefaultAgent struct{}

func (a *DefaultAgent) ChatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return ChatSession(model, msgs, pol)
}
