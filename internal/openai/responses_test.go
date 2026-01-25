package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

func TestResponsesClient_ChatOnce(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Errorf("expected /v1/responses, got %s", r.URL.Path)
		}

		var req responsesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		w.WriteHeader(200)
		resp := responsesResponse{
			ID: "resp_123",
			Output: []responsesOutputItem{
				{
					Type: "message",
					Message: &outputMessage{
						Role: "assistant",
						Content: []outputContent{
							{Type: "output_text", Text: "Hello from responses API"},
						},
					},
				},
			},
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	prev := os.Getenv("OPENAI_BASE_URL")
	_ = os.Setenv("OPENAI_BASE_URL", srv.URL)
	defer func() { _ = os.Setenv("OPENAI_BASE_URL", prev) }()

	client := responsesClient{}
	msgs := []types.Message{{Role: "user", Content: "hi"}}
	resp, err := client.ChatOnce("gpt-4o", msgs, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}

	if resp.ID != "resp_123" {
		t.Errorf("expected ID resp_123, got %s", resp.ID)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "Hello from responses API" {
		t.Errorf("unexpected content: %s", resp.Choices[0].Message.Content)
	}
}

func TestResponsesClient_ChatOnce_PreviousResponseID(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		var req responsesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.PreviousResponseID != "prev_123" {
			t.Errorf("expected PreviousResponseID prev_123, got %s", req.PreviousResponseID)
		}
		// input should only contain the NEW message (the one after the assistant message)
		if len(req.Input) != 1 {
			t.Errorf("expected 1 input item, got %d", len(req.Input))
		}

		w.WriteHeader(200)
		resp := responsesResponse{ID: "resp_456", Output: []responsesOutputItem{{Type: "message", Message: &outputMessage{Role: "assistant", Content: []outputContent{{Type: "text", Text: "next"}}}}}}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	prev := os.Getenv("OPENAI_BASE_URL")
	_ = os.Setenv("OPENAI_BASE_URL", srv.URL)
	defer func() { _ = os.Setenv("OPENAI_BASE_URL", prev) }()

	client := responsesClient{}
	msgs := []types.Message{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello", ResponseID: "prev_123"},
		{Role: "user", Content: "how are you?"},
	}
	_, err := client.ChatOnce("gpt-4o", msgs, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}
}

func TestResponsesClient_ChatOnce_InputStructure(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Input []struct {
				Type    string `json:"type"`
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if len(req.Input) != 1 {
			t.Errorf("expected 1 input item, got %d", len(req.Input))
		}
		item := req.Input[0]
		if item.Type != "message" {
			t.Errorf("expected type message, got %s", item.Type)
		}
		if item.Role != "user" {
			t.Errorf("expected role user, got %s", item.Role)
		}
		if len(item.Content) != 1 {
			t.Errorf("expected 1 content item, got %d", len(item.Content))
		}
		if item.Content[0].Type != "input_text" {
			t.Errorf("expected type input_text, got %s", item.Content[0].Type)
		}
		if item.Content[0].Text != "hi" {
			t.Errorf("expected text hi, got %s", item.Content[0].Text)
		}

		w.WriteHeader(200)
		resp := responsesResponse{ID: "resp_123", Output: []responsesOutputItem{{Type: "message", Message: &outputMessage{Role: "assistant", Content: []outputContent{{Type: "text", Text: "ok"}}}}}}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	prev := os.Getenv("OPENAI_BASE_URL")
	_ = os.Setenv("OPENAI_BASE_URL", srv.URL)
	defer func() { _ = os.Setenv("OPENAI_BASE_URL", prev) }()

	client := responsesClient{}
	msgs := []types.Message{{Role: "user", Content: "hi"}}
	_, err := client.ChatOnce("gpt-4o", msgs, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}
}

func TestResponsesClient_ChatOnce_ToolCall(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		var req responsesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		w.WriteHeader(200)
		resp := responsesResponse{
			ID: "resp_tool",
			Output: []responsesOutputItem{
				{
					Type:      "function_call",
					CallID:    "call_abc",
					Name:      "get_weather",
					Arguments: json.RawMessage(`{"location":"London"}`),
				},
			},
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	prev := os.Getenv("OPENAI_BASE_URL")
	_ = os.Setenv("OPENAI_BASE_URL", srv.URL)
	defer func() { _ = os.Setenv("OPENAI_BASE_URL", prev) }()

	client := responsesClient{}
	resp, err := client.ChatOnce("gpt-4o", nil, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}

	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %+v", resp.Choices[0].Message.ToolCalls)
	}
	tc := resp.Choices[0].Message.ToolCalls[0]
	if tc.Function.Name != "get_weather" {
		t.Errorf("expected get_weather, got %s", tc.Function.Name)
	}
}

func TestResponsesClient_ChatOnce_ToolMapping(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		var req responsesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if len(req.Tools) != 1 {
			t.Errorf("expected 1 tool, got %d", len(req.Tools))
		}

		// Check the first tool's structure
		toolMap, ok := req.Tools[0].(map[string]any)
		if !ok {
			t.Fatalf("expected tool to be a map, got %T", req.Tools[0])
		}
		if toolMap["type"] != "function" {
			t.Errorf("expected tool type function, got %v", toolMap["type"])
		}
		if toolMap["name"] != "test_tool" {
			t.Errorf("expected tool name test_tool, got %v", toolMap["name"])
		}

		w.WriteHeader(200)
		resp := responsesResponse{ID: "resp_123", Output: []responsesOutputItem{{Type: "message", Message: &outputMessage{Role: "assistant", Content: []outputContent{{Type: "text", Text: "ok"}}}}}}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	prev := os.Getenv("OPENAI_BASE_URL")
	_ = os.Setenv("OPENAI_BASE_URL", srv.URL)
	defer func() { _ = os.Setenv("OPENAI_BASE_URL", prev) }()

	client := responsesClient{}
	toolsList := []types.Tool{
		{
			Type: "function",
			Function: types.ToolFunction{
				Name: "test_tool",
			},
		},
	}
	_, err := client.ChatOnce("gpt-4o", nil, toolsList)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}
}

func TestResponsesClient_ChatOnce_ToolOutput_InputStructure(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Input []struct {
				Type   string `json:"type"`
				CallID string `json:"call_id"`
				Output string `json:"output"`
			} `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if len(req.Input) != 1 {
			t.Errorf("expected 1 input item, got %d", len(req.Input))
		}
		item := req.Input[0]
		if item.Type != "function_call_output" {
			t.Errorf("expected type function_call_output, got %s", item.Type)
		}
		if item.CallID != "call_abc" {
			t.Errorf("expected call_id call_abc, got %s", item.CallID)
		}
		if item.Output != "tool result" {
			t.Errorf("expected output 'tool result', got %s", item.Output)
		}

		w.WriteHeader(200)
		resp := responsesResponse{ID: "resp_123", Output: []responsesOutputItem{{Type: "message", Message: &outputMessage{Role: "assistant", Content: []outputContent{{Type: "text", Text: "ok"}}}}}}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	prev := os.Getenv("OPENAI_BASE_URL")
	_ = os.Setenv("OPENAI_BASE_URL", srv.URL)
	defer func() { _ = os.Setenv("OPENAI_BASE_URL", prev) }()

	client := responsesClient{}
	msgs := []types.Message{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "thinking", ResponseID: "prev_123"},
		{Role: "tool", Content: "tool result", ToolCallID: "call_abc"},
	}
	_, err := client.ChatOnce("gpt-4o", msgs, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}
}
