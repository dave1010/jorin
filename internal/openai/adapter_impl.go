package openai

import (
	"context"
	"github.com/dave1010/jorin/internal/types"
)

// DefaultAgent implements agent.Agent by delegating to package-level
// ChatSession and accepting a context for cancellation.
type DefaultAgent struct{}

func (a *DefaultAgent) ChatSession(ctx context.Context, model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return ChatSession(ctx, model, msgs, pol)
}
