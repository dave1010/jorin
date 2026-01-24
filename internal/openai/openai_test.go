package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

func TestChatOnceDecodesResponse(t *testing.T) {
	prevUseResponses := UseResponses
	UseResponses = true
	t.Cleanup(func() { UseResponses = prevUseResponses })

	// create a fake server returning a simple responses payload
	h := func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/responses") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		resp := types.Response{
			Output: []types.ResponseItem{
				{
					Type: "message",
					Role: "assistant",
					Content: []types.ResponseContent{
						{Type: "output_text", Text: "ok"},
					},
				},
			},
		}
		b, _ := json.Marshal(resp)
		if _, err := w.Write(b); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	// set OPENAI_BASE_URL to test server URL
	prev := os.Getenv("OPENAI_BASE_URL")
	if err := os.Setenv("OPENAI_BASE_URL", srv.URL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Setenv("OPENAI_BASE_URL", prev); err != nil {
			t.Fatalf("failed to restore env: %v", err)
		}
	}()

	msgs := []types.Message{{Role: "system", Content: "x"}}
	resp, err := ChatOnce("model", msgs, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "ok" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}

func TestChatOnceDecodesCompletionsResponse(t *testing.T) {
	prevUseResponses := UseResponses
	UseResponses = false
	t.Cleanup(func() { UseResponses = prevUseResponses })

	h := func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/chat/completions") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		resp := types.ChatResponse{Choices: []types.Choice{{Message: types.Message{Role: "assistant", Content: "ok"}}}}
		b, _ := json.Marshal(resp)
		if _, err := w.Write(b); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	prev := os.Getenv("OPENAI_BASE_URL")
	if err := os.Setenv("OPENAI_BASE_URL", srv.URL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Setenv("OPENAI_BASE_URL", prev); err != nil {
			t.Fatalf("failed to restore env: %v", err)
		}
	}()

	msgs := []types.Message{{Role: "system", Content: "x"}}
	resp, err := ChatOnce("model", msgs, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "ok" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}
