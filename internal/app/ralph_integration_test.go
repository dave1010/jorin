package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/prompt"
	"github.com/dave1010/jorin/internal/types"
)

func TestRunRalphLoopIntegration(t *testing.T) {
	prompt.EnableRalph()
	t.Cleanup(prompt.DisableRalph)

	llm := &recordingLLM{
		response: func(msgs []types.Message) types.ChatResponse {
			call := strings.TrimSpace(msgs[len(msgs)-1].Content)
			content := "still working"
			if call == "still working" {
				content = "DONE"
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
		Model:         "test-model",
		Prompt:        "start ralph",
		Policy:        types.Policy{},
		Stdin:         strings.NewReader(""),
		Stdout:        &stdout,
		Stderr:        &stderr,
		RalphMaxTries: 3,
	}

	if err := NewApp(&cfg).Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if llm.Calls() != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", llm.Calls())
	}

	msgs := llm.Messages()
	if len(msgs) != 2 || len(msgs[0]) < 2 {
		t.Fatalf("expected two calls with messages, got %#v", msgs)
	}
	if msgs[0][0].Role != "system" || !strings.Contains(msgs[0][0].Content, "Ralph Wiggum") {
		t.Fatalf("expected system prompt to include Ralph instructions")
	}
	if msgs[0][1].Content != "start ralph" {
		t.Fatalf("expected initial prompt, got %q", msgs[0][1].Content)
	}
	if msgs[1][1].Content != "still working" {
		t.Fatalf("expected second prompt to use previous output, got %q", msgs[1][1].Content)
	}

	if !strings.Contains(stdout.String(), "still working") || !strings.Contains(stdout.String(), "DONE") {
		t.Fatalf("expected stdout to include loop outputs, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Ralph iteration 1/3") || !strings.Contains(stderr.String(), "Ralph iteration 2/3") {
		t.Fatalf("expected stderr to include iteration logs, got %q", stderr.String())
	}
}
