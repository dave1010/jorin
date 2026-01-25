package openai

import "github.com/dave1010/jorin/internal/types"

// DefaultAgent implements types.Agent by delegating to package-level
// ChatSession.
type DefaultAgent struct {
	LLM LLM
}

func NewDefaultAgent(useResponsesAPI bool) *DefaultAgent {
	if useResponsesAPI {
		return &DefaultAgent{LLM: responsesClient{}}
	}
	return &DefaultAgent{LLM: nil}
}

func (a *DefaultAgent) ChatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	if a.LLM != nil {
		return chatSessionWithLLM(a.LLM, model, msgs, pol)
	}
	return ChatSession(model, msgs, pol)
}
