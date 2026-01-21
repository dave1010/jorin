package tools

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dave1010/jorin/internal/types"
)

func TestHTTPGet(t *testing.T) {
	r := Registry()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("pong")); err != nil {
			return
		}
	}))
	defer srv.Close()

	out, err := r["http_get"](map[string]any{"url": srv.URL}, &types.Policy{})
	if err != nil {
		t.Fatalf("http_get failed: %v", err)
	}
	if body, _ := out["body"].(string); body != "pong" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestHTTPGetWaitsForResponse(t *testing.T) {
	r := Registry()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		if _, err := w.Write([]byte("late")); err != nil {
			return
		}
	}))
	defer srv.Close()

	start := time.Now()
	out, err := r["http_get"](map[string]any{"url": srv.URL}, &types.Policy{})
	if err != nil {
		t.Fatalf("http_get delayed failed: %v", err)
	}
	dur := time.Since(start)
	if body, _ := out["body"].(string); body != "late" {
		t.Fatalf("unexpected body delayed: %q", body)
	}
	if dur < 0 || dur > 5*time.Second {
		t.Fatalf("unexpected duration: %v", dur)
	}
}
