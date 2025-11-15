package agent

import (
	"context"
	"github.com/dave1010/jorin/internal/types"
)

// Agent is a minimal interface for an LLM backend used by the UI.
// Implementations should provide ChatSession similar to the previous
// package-level function but accept a context so calls can be cancelled.
type Agent interface {
	ChatSession(ctx context.Context, model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error)
}
