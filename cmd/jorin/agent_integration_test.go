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

func newOpenAIServer(t *testing.T, handler func(t *testing.T, req types.ChatRequest, call int) types.ChatResponse) *openAIServer {
	t.Helper()

	ois := &openAIServer{}
	ois.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		ois.mu.Lock()
		ois.count++
		current := ois.count
		ois.mu.Unlock()

		if !strings.HasSuffix(r.URL.Path, "/v1/chat/completions") {
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

	openAIServer := newOpenAIServer(t, func(t *testing.T, req types.ChatRequest, current int) types.ChatResponse {
		switch current {
		case 1:
			if len(req.Messages) != 2 {
				t.Errorf("expected 2 messages, got %d", len(req.Messages))
			}
			if len(req.Messages) > 0 {
				if req.Messages[0].Role != "system" || !strings.Contains(req.Messages[0].Content, "You are Jorin") {
					t.Errorf("expected system prompt in first message")
				}
			}
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role: "assistant",
							ToolCalls: []types.ToolCall{
								{
									ID:   "call_read",
									Type: "function",
									Function: struct {
										Name string          `json:"name"`
										Args json.RawMessage `json:"arguments"`
									}{
										Name: "read_file",
										Args: json.RawMessage(`{"path":"` + filePath + `"}`),
									},
								},
								{
									ID:   "call_http",
									Type: "function",
									Function: struct {
										Name string          `json:"name"`
										Args json.RawMessage `json:"arguments"`
									}{
										Name: "http_get",
										Args: json.RawMessage(`{"url":"` + httpServer.URL + `"}`),
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
			}
		case 2:
			toolMessages := []types.Message{}
			for _, msg := range req.Messages {
				if msg.Role == "tool" {
					toolMessages = append(toolMessages, msg)
				}
			}
			if len(toolMessages) != 2 {
				t.Errorf("expected 2 tool messages, got %d", len(toolMessages))
			}
			for _, msg := range toolMessages {
				var payload map[string]any
				if err := json.Unmarshal([]byte(msg.Content), &payload); err != nil {
					t.Errorf("decode tool payload: %v", err)
					continue
				}
				switch msg.Name {
				case "read_file":
					if payload["text"] != "hello from file" {
						t.Errorf("expected read_file payload, got %v", payload["text"])
					}
				case "http_get":
					if payload["status"] != float64(http.StatusOK) || payload["body"] != "payload" {
						t.Errorf("expected http_get payload, got %v", payload)
					}
				default:
					t.Errorf("unexpected tool name: %s", msg.Name)
				}
			}
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role:    "assistant",
							Content: "done",
						},
						FinishReason: "stop",
					},
				},
			}
		default:
			t.Fatalf("unexpected request: %d", current)
		}

		return types.ChatResponse{}
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

	openAIServer := newOpenAIServer(t, func(t *testing.T, req types.ChatRequest, current int) types.ChatResponse {
		switch current {
		case 1:
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role: "assistant",
							ToolCalls: []types.ToolCall{
								{
									ID:   "call_read",
									Type: "function",
									Function: struct {
										Name string          `json:"name"`
										Args json.RawMessage `json:"arguments"`
									}{
										Name: "read_file",
										Args: json.RawMessage(strconv.Quote(filePath)),
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
			}
		case 2:
			toolMessages := []types.Message{}
			for _, msg := range req.Messages {
				if msg.Role == "tool" {
					toolMessages = append(toolMessages, msg)
				}
			}
			if len(toolMessages) != 1 {
				t.Errorf("expected 1 tool message, got %d", len(toolMessages))
			}
			if len(toolMessages) == 1 {
				var payload map[string]any
				if err := json.Unmarshal([]byte(toolMessages[0].Content), &payload); err != nil {
					t.Errorf("decode tool payload: %v", err)
				} else if payload["text"] != "string args" {
					t.Errorf("expected string args payload, got %v", payload["text"])
				}
			}
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role:    "assistant",
							Content: "done",
						},
						FinishReason: "stop",
					},
				},
			}
		default:
			t.Fatalf("unexpected request: %d", current)
		}
		return types.ChatResponse{}
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
	openAIServer := newOpenAIServer(t, func(t *testing.T, req types.ChatRequest, current int) types.ChatResponse {
		switch current {
		case 1:
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role: "assistant",
							ToolCalls: []types.ToolCall{
								{
									ID:   "call_unknown",
									Type: "function",
									Function: struct {
										Name string          `json:"name"`
										Args json.RawMessage `json:"arguments"`
									}{
										Name: "mystery_tool",
										Args: json.RawMessage(`{"topic":"tests"}`),
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
			}
		case 2:
			toolMessages := []types.Message{}
			for _, msg := range req.Messages {
				if msg.Role == "tool" {
					toolMessages = append(toolMessages, msg)
				}
			}
			if len(toolMessages) != 1 {
				t.Errorf("expected 1 tool message, got %d", len(toolMessages))
			}
			if len(toolMessages) == 1 {
				if toolMessages[0].Name != "mystery_tool" {
					t.Errorf("expected mystery_tool name, got %s", toolMessages[0].Name)
				}
				var payload map[string]any
				if err := json.Unmarshal([]byte(toolMessages[0].Content), &payload); err != nil {
					t.Errorf("decode tool payload: %v", err)
				} else if payload["error"] != "unknown tool" {
					t.Errorf("expected unknown tool payload, got %v", payload["error"])
				}
			}
			return types.ChatResponse{
				Choices: []types.Choice{
					{
						Message: types.Message{
							Role:    "assistant",
							Content: "done",
						},
						FinishReason: "stop",
					},
				},
			}
		default:
			t.Fatalf("unexpected request: %d", current)
		}
		return types.ChatResponse{}
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
