package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

func TestRunAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var resp types.ChatResponse
		if len(req.Messages) == 2 && req.Messages[1].Role == "user" {
			resp = types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role: "assistant",
							ToolCalls: []types.ToolCall{
								{
									ID:   "call_123",
									Type: "function",
									Function: struct {
										Name string `json:"name"`
										Args string `json:"arguments"`
									}{
										Name: "read_file",
										Args: `{"path":"test.txt"}`,
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
			}
		} else if len(req.Messages) > 2 && req.Messages[len(req.Messages)-1].Role == "tool" {
			resp = types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role:    "assistant",
							Content: "File content: hello",
						},
						FinishReason: "stop",
					},
				},
			}
		} else {
			t.Errorf("Unexpected request: %v", req)
			http.Error(w, "Unexpected request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	os.Setenv("OPENAI_BASE_URL", server.URL)
	os.Setenv("OPENAI_API_KEY", "test-key")

	if err := os.WriteFile("test.txt", []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove("test.txt")

	pol := &types.Policy{}
	out, err := runAgent("test-model", "read the file test.txt", pol)
	if err != nil {
		t.Fatalf("runAgent failed: %v", err)
	}

	if !strings.Contains(out, "File content: hello") {
		t.Errorf("Expected output to contain 'File content: hello', but got: %s", out)
	}
}
