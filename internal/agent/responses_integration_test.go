package agent

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/openai"
	"github.com/dave1010/jorin/internal/types"
)

func TestResponsesAPIIntegration(t *testing.T) {
	openai.UseResponsesAPI()
	t.Cleanup(openai.UseCompletionsAPI)

	rs := newResponsesServer(t, func(t *testing.T, req map[string]any, current int) map[string]any {
		switch current {
		case 1:
			// Initial request
			if req["previous_response_id"] != nil && req["previous_response_id"] != "" {
				t.Errorf("expected no previous_response_id in first call, got %v", req["previous_response_id"])
			}
			input := req["input"].([]any)
			if len(input) != 1 {
				t.Errorf("expected 1 input item, got %d", len(input))
			}

			return map[string]any{
				"id": "resp_1",
				"output": []any{
					map[string]any{
						"type":      "function_call",
						"call_id":   "call_1",
						"name":      "read_file",
						"arguments": json.RawMessage(`{"path":"test.txt"}`),
					},
				},
			}
		case 2:
			// After tool result
			if req["previous_response_id"] != "resp_1" {
				t.Errorf("expected previous_response_id resp_1, got %v", req["previous_response_id"])
			}
			input := req["input"].([]any)
			// Jorin should only send the NEW input (the tool result)
			if len(input) != 1 {
				t.Errorf("expected 1 input item (tool output), got %d", len(input))
			}
			item := input[0].(map[string]any)
			if item["type"] != "function_call_output" {
				t.Errorf("expected function_call_output, got %v", item["type"])
			}
			if item["call_id"] != "call_1" {
				t.Errorf("expected call_id call_1, got %v", item["call_id"])
			}

			return map[string]any{
				"id": "resp_2",
				"output": []any{
					map[string]any{
						"type": "message",
						"message": map[string]any{
							"role": "assistant",
							"content": []any{
								map[string]any{
									"type": "output_text",
									"text": "file content is hello",
								},
							},
						},
					},
				},
			}
		default:
			t.Fatalf("unexpected call %d", current)
			return nil
		}
	})
	defer rs.Close()

	t.Setenv("OPENAI_BASE_URL", rs.URL())
	t.Setenv("OPENAI_API_KEY", "test-key")

	if err := os.WriteFile("test.txt", []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer func() { _ = os.Remove("test.txt") }()

	pol := &types.Policy{}
	out, err := RunAgent("test-model", "read test.txt", "you are jorin", pol)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}

	if !strings.Contains(out, "file content is hello") {
		t.Errorf("expected output to contain 'file content is hello', got %q", out)
	}

	if rs.Count() != 2 {
		t.Errorf("expected 2 calls to responses API, got %d", rs.Count())
	}
}
