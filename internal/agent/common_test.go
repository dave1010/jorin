package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

type responsesServer struct {
	server *httptest.Server
	mu     sync.Mutex
	count  int
}

func newResponsesServer(t *testing.T, handler func(t *testing.T, req map[string]any, call int) map[string]any) *responsesServer {
	t.Helper()

	rs := &responsesServer{}
	rs.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		rs.mu.Lock()
		rs.count++
		current := rs.count
		rs.mu.Unlock()

		if !strings.HasSuffix(r.URL.Path, "/v1/responses") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := handler(t, req, current)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))

	return rs
}

func (r *responsesServer) Close() {
	r.server.Close()
}

func (r *responsesServer) URL() string {
	return r.server.URL
}

func (r *responsesServer) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}
