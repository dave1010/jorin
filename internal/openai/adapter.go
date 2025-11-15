package openai

import (
	"context"
	"github.com/dave1010/jorin/internal/types"
)

// Ensure default LLM implements the Agent interface shape used by UI/agent
// code. We provide an adapter function to satisfy agent.Agent if needed.

// Adapter wraps the package-level ChatSession to match the agent.Agent
// interface shape. Note: we accept a context so callers can cancel requests.
func Adapter(ctx context.Context, model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return ChatSession(ctx, model, msgs, pol)
}
