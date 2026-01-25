package app

import (
	"sync"
	"testing"

	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
)

type recordingLLM struct {
	mu       sync.Mutex
	calls    int
	messages [][]types.Message
	response func(msgs []types.Message) types.ChatResponse
}

func (r *recordingLLM) ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	r.mu.Lock()
	r.calls++
	snapshot := append([]types.Message(nil), msgs...)
	r.messages = append(r.messages, snapshot)
	r.mu.Unlock()

	resp := types.ChatResponse{
		Choices: []types.Choice{
			{
				Message: types.Message{
					Role:    "assistant",
					Content: "ok",
				},
				FinishReason: "stop",
			},
		},
	}
	if r.response != nil {
		resp = r.response(msgs)
	}
	return &resp, nil
}

func (r *recordingLLM) Calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func (r *recordingLLM) Messages() [][]types.Message {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([][]types.Message, len(r.messages))
	for i, msgs := range r.messages {
		out[i] = append([]types.Message(nil), msgs...)
	}
	return out
}

func withTestLLM(t *testing.T, llm openai.LLM) {
	t.Helper()
	orig := openai.DefaultLLM
	openai.DefaultLLM = llm
	t.Cleanup(func() {
		openai.DefaultLLM = orig
	})
}
