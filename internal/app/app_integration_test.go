package app

import (
	"bytes"
	"context"
	"strings"
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

func TestRunPromptWritesOutput(t *testing.T) {
	llm := &recordingLLM{
		response: func(msgs []types.Message) types.ChatResponse {
			content := "reply"
			if len(msgs) > 0 {
				last := msgs[len(msgs)-1]
				content = "reply: " + last.Content
			}
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role:    "assistant",
							Content: content,
						},
						FinishReason: "stop",
					},
				},
			}
		},
	}
	withTestLLM(t, llm)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	opts := Options{
		Model:  "test-model",
		Prompt: "say hi",
		Policy: types.Policy{},
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	if err := Run(context.Background(), opts); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "reply: say hi") {
		t.Fatalf("expected output to include reply, got %q", stdout.String())
	}

	if llm.Calls() != 1 {
		t.Fatalf("expected 1 LLM call, got %d", llm.Calls())
	}

	msgs := llm.Messages()
	if len(msgs) != 1 || len(msgs[0]) != 2 {
		t.Fatalf("expected 1 call with 2 messages, got %#v", msgs)
	}
	if msgs[0][0].Role != "system" {
		t.Fatalf("expected system message first, got %s", msgs[0][0].Role)
	}
	if msgs[0][1].Role != "user" || msgs[0][1].Content != "say hi" {
		t.Fatalf("expected user prompt, got %+v", msgs[0][1])
	}
}

func TestRunMissingPrompt(t *testing.T) {
	llm := &recordingLLM{}
	withTestLLM(t, llm)

	opts := Options{
		Model:  "test-model",
		Prompt: "   ",
		Policy: types.Policy{},
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	err := Run(context.Background(), opts)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrMissingPrompt {
		t.Fatalf("expected ErrMissingPrompt, got %v", err)
	}
	if llm.Calls() != 0 {
		t.Fatalf("expected no LLM calls, got %d", llm.Calls())
	}
}

func TestRunREPLCommandsAndHistory(t *testing.T) {
	llm := &recordingLLM{
		response: func(msgs []types.Message) types.ChatResponse {
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role:    "assistant",
							Content: "ack",
						},
						FinishReason: "stop",
					},
				},
			}
		},
	}
	withTestLLM(t, llm)

	input := strings.NewReader("hello\n/history\n/help repl\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	opts := Options{
		Model:  "test-model",
		Prompt: "ignored",
		Repl:   true,
		Policy: types.Policy{},
		Stdin:  input,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	if err := Run(context.Background(), opts); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if llm.Calls() != 1 {
		t.Fatalf("expected 1 LLM call, got %d", llm.Calls())
	}

	msgs := llm.Messages()
	if len(msgs) != 1 || len(msgs[0]) < 2 {
		t.Fatalf("expected messages to include user input, got %#v", msgs)
	}
	if msgs[0][1].Content != "hello" {
		t.Fatalf("expected user message to be hello, got %q", msgs[0][1].Content)
	}

	out := stdout.String()
	if !strings.Contains(out, "ack") {
		t.Fatalf("expected response output, got %q", out)
	}
	if !strings.Contains(out, "repl - editing tips") {
		t.Fatalf("expected help output, got %q", out)
	}
	if !strings.Contains(out, "> hello\n") {
		t.Fatalf("expected history output to include hello, got %q", out)
	}
}
