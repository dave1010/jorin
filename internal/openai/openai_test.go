package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

func TestChatOnceDecodesResponse(t *testing.T) {
	// create a fake server returning a simple chat response
	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		resp := types.ChatResponse{Choices: []types.Choice{{Message: types.Message{Role: "assistant", Content: "ok"}}}}
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	// set OPENAI_BASE_URL to test server URL
	prev := os.Getenv("OPENAI_BASE_URL")
	defer os.Setenv("OPENAI_BASE_URL", prev)
	os.Setenv("OPENAI_BASE_URL", srv.URL)

	msgs := []types.Message{{Role: "system", Content: "x"}}
	resp, err := ChatOnce("model", msgs, nil)
	if err != nil {
		t.Fatalf("ChatOnce error: %v", err)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "ok" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}
