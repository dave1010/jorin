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
		if !strings.HasSuffix(r.URL.Path, "/v1/responses") {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var req struct {
			Input []types.ResponseItem `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var resp types.Response
		if len(req.Input) == 2 && req.Input[1].Type == "message" && req.Input[1].Role == "user" {
			resp = types.Response{
				Output: []types.ResponseItem{
					{
						Type:      "function_call",
						Name:      "read_file",
						Arguments: json.RawMessage(`{"path":"test.txt"}`),
						CallID:    "call_123",
					},
				},
			}
		} else if len(req.Input) > 2 && req.Input[len(req.Input)-1].Type == "function_call_output" {
			resp = types.Response{
				Output: []types.ResponseItem{
					{
						Type: "message",
						Role: "assistant",
						Content: []types.ResponseContent{
							{Type: "output_text", Text: "File content: hello"},
						},
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

	if err := os.Setenv("OPENAI_BASE_URL", server.URL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("OPENAI_API_KEY", "test-key"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	if err := os.WriteFile("test.txt", []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer func() {
		if err := os.Remove("test.txt"); err != nil {
			t.Fatalf("failed to remove test file: %v", err)
		}
	}()

	pol := &types.Policy{}
	out, err := runAgent("test-model", "read the file test.txt", pol)
	if err != nil {
		t.Fatalf("runAgent failed: %v", err)
	}

	if !strings.Contains(out, "File content: hello") {
		t.Errorf("Expected output to contain 'File content: hello', but got: %s", out)
	}
}
