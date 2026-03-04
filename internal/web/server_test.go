package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

type fakeAgent struct {
	out      string
	err      error
	calls    int
	lastMsgs []types.Message
}

func (f *fakeAgent) ChatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	f.calls++
	f.lastMsgs = append([]types.Message(nil), msgs...)
	return nil, f.out, f.err
}

func TestHandleIndex(t *testing.T) {
	srv := NewServer(Config{Agent: &fakeAgent{}, Model: "test"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	srv.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Jorin Web Service") {
		t.Fatalf("expected page title content, got %q", rr.Body.String())
	}
}

func TestHandleChat(t *testing.T) {
	agent := &fakeAgent{out: "hello from jorin"}
	srv := NewServer(Config{Agent: agent, Model: "test"})

	body := map[string]any{
		"messages": []map[string]string{{"role": "user", "content": "say hi"}},
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(data))
	rr := httptest.NewRecorder()
	srv.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if agent.calls != 1 {
		t.Fatalf("expected one agent call, got %d", agent.calls)
	}
	if len(agent.lastMsgs) < 2 || agent.lastMsgs[1].Content != "say hi" {
		t.Fatalf("expected forwarded prompt, got %#v", agent.lastMsgs)
	}

	var resp chatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if resp.Reply != "hello from jorin" {
		t.Fatalf("expected reply, got %q", resp.Reply)
	}
}

func TestHandleChatRejectsInvalidRole(t *testing.T) {
	srv := NewServer(Config{Agent: &fakeAgent{out: "ok"}, Model: "test"})

	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"messages":[{"role":"system","content":"nope"}]}`))
	rr := httptest.NewRecorder()
	srv.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
