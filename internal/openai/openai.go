package openai

import (
	"errors"

	"github.com/dave1010/jorin/internal/tools"
	"github.com/dave1010/jorin/internal/types"
)

const maxChatTurns = 100

// ChatOnce is a convenience wrapper that delegates to the package-level
// DefaultLLM implementation. Callers can swap DefaultLLM for a different
// provider in tests or to support other LLMs.
func ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	return DefaultLLM.ChatOnce(model, msgs, toolsList)
}

func ChatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	return chatSessionWithLLM(DefaultLLM, model, msgs, pol)
}

func chatSessionWithLLM(llm LLM, model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	toolsList := tools.ToolsManifest()
	reg := tools.Registry()
	for i := 0; i < maxChatTurns; i++ {
		resp, err := llm.ChatOnce(model, msgs, toolsList)
		if err != nil {
			return msgs, "", err
		}
		if len(resp.Choices) == 0 {
			return msgs, "", errors.New("no choices")
		}
		ch := resp.Choices[0]
		cm := ch.Message
		cm.ResponseID = resp.ID

		msgs = append(msgs, cm)

		if len(cm.ToolCalls) > 0 {
			toolMsgs := handleToolCalls(cm.ToolCalls, reg, pol)
			msgs = append(msgs, toolMsgs...)
			continue
		}

		return msgs, cm.Content, nil
	}
	return msgs, "", errors.New("max turns reached")
}
