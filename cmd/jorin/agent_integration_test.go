package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

type openAIServer struct {
	server *httptest.Server
	mu     sync.Mutex
	count  int
}

type responseRequest struct {
	Model string               `json:"model"`
	Input []types.ResponseItem `json:"input"`
}

func newOpenAIServer(t *testing.T, handler func(t *testing.T, req responseRequest, call int) types.Response) *openAIServer {
	t.Helper()

	ois := &openAIServer{}
	ois.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req responseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		ois.mu.Lock()
		ois.count++
		current := ois.count
		ois.mu.Unlock()

		if !strings.HasSuffix(r.URL.Path, "/v1/responses") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := handler(t, req, current)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))

	return ois
}

func responseInputText(item types.ResponseItem) string {
	for _, content := range item.Content {
		if content.Type == "input_text" {
			return content.Text
		}
	}
	return ""
}

func responseMessageItems(items []types.ResponseItem) []types.ResponseItem {
	var msgs []types.ResponseItem
	for _, item := range items {
		if item.Type == "message" {
			msgs = append(msgs, item)
		}
	}
	return msgs
}

func responseToolOutputs(items []types.ResponseItem) []types.ResponseItem {
	var outputs []types.ResponseItem
	for _, item := range items {
		if item.Type == "function_call_output" {
			outputs = append(outputs, item)
		}
	}
	return outputs
}

func (o *openAIServer) Close() {
	o.server.Close()
}

func (o *openAIServer) URL() string {
	return o.server.URL
}

func (o *openAIServer) Count() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.count
}

func TestRunAgentIntegrationFlow(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "note.txt")
	if err := os.WriteFile(filePath, []byte("hello from file"), 0o644); err != nil {
		t.Fatalf("write note.txt: %v", err)
	}

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("payload"))
	}))
	defer httpServer.Close()

	openAIServer := newOpenAIServer(t, func(t *testing.T, req responseRequest, current int) types.Response {
		switch current {
		case 1:
			messages := responseMessageItems(req.Input)
			if len(messages) != 2 {
				t.Errorf("expected 2 messages, got %d", len(messages))
			}
			if len(messages) > 0 {
				if messages[0].Role != "system" || !strings.Contains(responseInputText(messages[0]), "You are Jorin") {
					t.Errorf("expected system prompt in first message")
				}
			}
			return types.Response{
				Output: []types.ResponseItem{
					{
						Type:      "function_call",
						Name:      "read_file",
						Arguments: json.RawMessage(`{"path":"` + filePath + `"}`),
						CallID:    "call_read",
					},
					{
						Type:      "function_call",
						Name:      "http_get",
						Arguments: json.RawMessage(`{"url":"` + httpServer.URL + `"}`),
						CallID:    "call_http",
					},
				},
			}
		case 2:
			toolOutputs := responseToolOutputs(req.Input)
			if len(toolOutputs) != 2 {
				t.Errorf("expected 2 tool outputs, got %d", len(toolOutputs))
			}
			for _, item := range toolOutputs {
				var payload map[string]any
				if err := json.Unmarshal([]byte(item.Output), &payload); err != nil {
					t.Errorf("decode tool payload: %v", err)
					continue
				}
				switch item.CallID {
				case "call_read":
					if payload["text"] != "hello from file" {
						t.Errorf("expected read_file payload, got %v", payload["text"])
					}
				case "call_http":
					if payload["status"] != float64(http.StatusOK) || payload["body"] != "payload" {
						t.Errorf("expected http_get payload, got %v", payload)
					}
				default:
					t.Errorf("unexpected tool call id: %s", item.CallID)
				}
			}
			return types.Response{
				Output: []types.ResponseItem{
					{
						Type: "message",
						Role: "assistant",
						Content: []types.ResponseContent{
							{Type: "output_text", Text: "done"},
						},
					},
				},
			}
		default:
			t.Fatalf("unexpected request: %d", current)
		}

		return types.Response{}
	})
	defer openAIServer.Close()

	t.Setenv("OPENAI_BASE_URL", openAIServer.URL())
	t.Setenv("OPENAI_API_KEY", "test-key")

	pol := &types.Policy{}
	out, err := runAgent("test-model", "read and fetch", pol)
	if err != nil {
		t.Fatalf("runAgent failed: %v", err)
	}
	if out != "done" {
		t.Fatalf("expected output to be done, got: %s", out)
	}

	if openAIServer.Count() != 2 {
		t.Fatalf("expected 2 requests, got %d", openAIServer.Count())
	}
}

func TestRunAgentIntegrationStringArgsFallback(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "note.txt")
	if err := os.WriteFile(filePath, []byte("string args"), 0o644); err != nil {
		t.Fatalf("write note.txt: %v", err)
	}

	openAIServer := newOpenAIServer(t, func(t *testing.T, req responseRequest, current int) types.Response {
		switch current {
		case 1:
			return types.Response{
				Output: []types.ResponseItem{
					{
						Type:      "function_call",
						Name:      "read_file",
						Arguments: json.RawMessage(strconv.Quote(filePath)),
						CallID:    "call_read",
					},
				},
			}
		case 2:
			toolOutputs := responseToolOutputs(req.Input)
			if len(toolOutputs) != 1 {
				t.Errorf("expected 1 tool output, got %d", len(toolOutputs))
			}
			if len(toolOutputs) == 1 {
				var payload map[string]any
				if err := json.Unmarshal([]byte(toolOutputs[0].Output), &payload); err != nil {
					t.Errorf("decode tool payload: %v", err)
				} else if payload["text"] != "string args" {
					t.Errorf("expected string args payload, got %v", payload["text"])
				}
			}
			return types.Response{
				Output: []types.ResponseItem{
					{
						Type: "message",
						Role: "assistant",
						Content: []types.ResponseContent{
							{Type: "output_text", Text: "done"},
						},
					},
				},
			}
		default:
			t.Fatalf("unexpected request: %d", current)
		}
		return types.Response{}
	})
	defer openAIServer.Close()

	t.Setenv("OPENAI_BASE_URL", openAIServer.URL())
	t.Setenv("OPENAI_API_KEY", "test-key")

	pol := &types.Policy{}
	out, err := runAgent("test-model", "read file", pol)
	if err != nil {
		t.Fatalf("runAgent failed: %v", err)
	}
	if out != "done" {
		t.Fatalf("expected output to be done, got: %s", out)
	}

	if openAIServer.Count() != 2 {
		t.Fatalf("expected 2 requests, got %d", openAIServer.Count())
	}
}

func TestRunAgentIntegrationUnknownTool(t *testing.T) {
	openAIServer := newOpenAIServer(t, func(t *testing.T, req responseRequest, current int) types.Response {
		switch current {
		case 1:
			return types.Response{
				Output: []types.ResponseItem{
					{
						Type:      "function_call",
						Name:      "mystery_tool",
						Arguments: json.RawMessage(`{"topic":"tests"}`),
						CallID:    "call_unknown",
					},
				},
			}
		case 2:
			toolOutputs := responseToolOutputs(req.Input)
			if len(toolOutputs) != 1 {
				t.Errorf("expected 1 tool output, got %d", len(toolOutputs))
			}
			if len(toolOutputs) == 1 {
				if toolOutputs[0].CallID != "call_unknown" {
					t.Errorf("expected call_unknown id, got %s", toolOutputs[0].CallID)
				}
				var payload map[string]any
				if err := json.Unmarshal([]byte(toolOutputs[0].Output), &payload); err != nil {
					t.Errorf("decode tool payload: %v", err)
				} else if payload["error"] != "unknown tool" {
					t.Errorf("expected unknown tool payload, got %v", payload["error"])
				}
			}
			return types.Response{
				Output: []types.ResponseItem{
					{
						Type: "message",
						Role: "assistant",
						Content: []types.ResponseContent{
							{Type: "output_text", Text: "done"},
						},
					},
				},
			}
		default:
			t.Fatalf("unexpected request: %d", current)
		}
		return types.Response{}
	})
	defer openAIServer.Close()

	t.Setenv("OPENAI_BASE_URL", openAIServer.URL())
	t.Setenv("OPENAI_API_KEY", "test-key")

	pol := &types.Policy{}
	out, err := runAgent("test-model", "unknown tool", pol)
	if err != nil {
		t.Fatalf("runAgent failed: %v", err)
	}
	if out != "done" {
		t.Fatalf("expected output to be done, got: %s", out)
	}

	if openAIServer.Count() != 2 {
		t.Fatalf("expected 2 requests, got %d", openAIServer.Count())
	}
}
