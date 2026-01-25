package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

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
	cfg := Config{
		Model:  "test-model",
		Prompt: "say hi",
		Policy: types.Policy{},
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	if err := NewApp(&cfg).Run(context.Background()); err != nil {
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

	cfg := Config{
		Model:  "test-model",
		Prompt: "   ",
		Policy: types.Policy{},
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	err := NewApp(&cfg).Run(context.Background())
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

func TestRunPromptIncludesArgsAndStdin(t *testing.T) {
	llm := &recordingLLM{}
	withTestLLM(t, llm)

	var stdout bytes.Buffer
	cfg := Config{
		Model:      "test-model",
		Prompt:     "summarize",
		ScriptArgs: []string{"--format", "short"},
		Policy:     types.Policy{},
		Stdin:      strings.NewReader("doc text\n"),
		StdinIsTTY: false,
		Stdout:     &stdout,
		Stderr:     &bytes.Buffer{},
	}

	if err := NewApp(&cfg).Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	msgs := llm.Messages()
	if len(msgs) != 1 || len(msgs[0]) < 2 {
		t.Fatalf("expected messages to include prompt, got %#v", msgs)
	}

	expected := "summarize\n\nArguments: --format short\n\nStdin:\ndoc text"
	if msgs[0][1].Content != expected {
		t.Fatalf("expected prompt %q, got %q", expected, msgs[0][1].Content)
	}
}

func TestRunUsesStdinWhenNoPrompt(t *testing.T) {
	llm := &recordingLLM{}
	withTestLLM(t, llm)

	cfg := Config{
		Model:      "test-model",
		Prompt:     "",
		Policy:     types.Policy{},
		Stdin:      strings.NewReader("input only\n"),
		StdinIsTTY: false,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}

	if err := NewApp(&cfg).Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	msgs := llm.Messages()
	if len(msgs) != 1 || len(msgs[0]) < 2 {
		t.Fatalf("expected messages to include prompt, got %#v", msgs)
	}

	if msgs[0][1].Content != "input only" {
		t.Fatalf("expected stdin-only prompt, got %q", msgs[0][1].Content)
	}
}
