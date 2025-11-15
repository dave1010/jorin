package openai

import "github.com/dave1010/jorin/internal/types"

// Ensure default LLM implements the Agent interface shape used by UI/agent
// code. We provide an adapter function to satisfy types.Agent if needed.

// Adapter wraps the package-level ChatSession to match the types.Agent
// interface shape. Note: we won't add new dependencies here.
func Adapter(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return ChatSession(model, msgs, pol)
}
