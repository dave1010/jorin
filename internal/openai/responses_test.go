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

func TestResponsesClient_ChatOnce_ToolCall(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		resp := responsesResponse{
			ID: "resp_tool",
			Output: []responsesOutputItem{
				{
					Type:      "function_call",
					ID:        "call_abc",
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
