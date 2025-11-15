package agent

import "github.com/dave1010/jorin/internal/types"

// Agent is a minimal interface for an LLM backend used by the UI.
// Implementations should provide ChatSession similar to the previous
// package-level function.
type Agent interface {
	ChatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error)
}
