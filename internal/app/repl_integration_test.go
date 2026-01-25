package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

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
	cfg := Config{
		Model:  "test-model",
		Prompt: "ignored",
		Repl:   true,
		Policy: types.Policy{},
		Stdin:  input,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	if err := NewApp(&cfg).Run(context.Background()); err != nil {
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
