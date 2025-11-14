package tools

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dave1010/jorin/internal/types"
)

func TestReadWriteHttpAndDryShell(t *testing.T) {
	r := Registry()

	// read_file missing path
	if _, err := r["read_file"](map[string]any{}, &types.Policy{}); err == nil {
		t.Fatalf("expected error for missing path")
	}

	// create temp file and test read
	tmp := t.TempDir()
	f := filepath.Join(tmp, "t.txt")
	if err := os.WriteFile(f, []byte("hello tools"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	out, err := r["read_file"](map[string]any{"path": f}, &types.Policy{})
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}
	if out["text"] != "hello tools" {
		t.Fatalf("unexpected text: %#v", out["text"])
	}

	// write_file readonly
	wf := r["write_file"]
	outw, err := wf(map[string]any{"path": filepath.Join(tmp, "x.txt"), "text": "ok"}, &types.Policy{Readonly: true})
	if err != nil {
		t.Fatalf("write_file returned err: %v", err)
	}
	if outw["error"] != "readonly session" {
		t.Fatalf("expected readonly error, got: %#v", outw)
	}

	// write_file success
	outw, err = wf(map[string]any{"path": filepath.Join(tmp, "x.txt"), "text": "ok"}, &types.Policy{})
	if err != nil {
		t.Fatalf("write_file failed: %v", err)
	}
	if ok, _ := outw["ok"].(bool); !ok {
		t.Fatalf("expected ok true, got %#v", outw)
	}

	// http_get with test server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}))
	defer srv.Close()
	outH, err := r["http_get"](map[string]any{"url": srv.URL}, &types.Policy{})
	if err != nil {
		t.Fatalf("http_get failed: %v", err)
	}
	if body, _ := outH["body"].(string); body != "pong" {
		t.Fatalf("unexpected body: %q", body)
	}

	// dry shell
	outS, err := r["shell"](map[string]any{"cmd": "echo hi"}, &types.Policy{DryShell: true})
	if err != nil {
		t.Fatalf("shell dry failed: %v", err)
	}
	if dr, _ := outS["dry_run"].(bool); !dr {
		t.Fatalf("expected dry_run true")
	}
	if outS["cmd"].(string) != "echo hi" {
		t.Fatalf("unexpected cmd in dry run: %#v", outS)
	}

	// shell actual: use simple echo
	outS, err = r["shell"](map[string]any{"cmd": "echo -n OK"}, &types.Policy{})
	if err != nil {
		t.Fatalf("shell failed: %v", err)
	}
	if s, _ := outS["stdout"].(string); s != "OK" {
		t.Fatalf("unexpected stdout: %q", s)
	}

	// CWD test
	script := filepath.Join(tmp, "pw.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho -n $(pwd) > outpwd"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	outS, err = r["shell"](map[string]any{"cmd": "./pw.sh"}, &types.Policy{CWD: tmp})
	if err != nil {
		t.Fatalf("shell CWD failed: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(tmp, "outpwd"))
	if err != nil {
		t.Fatalf("read outpwd: %v", err)
	}
	if string(b) != tmp {
		t.Fatalf("expected %q got %q", tmp, string(b))
	}

	// http_get timeout shorter server
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte("late"))
	}))
	defer srv2.Close()
	start := time.Now()
	outH, err = r["http_get"](map[string]any{"url": srv2.URL}, &types.Policy{})
	if err != nil {
		t.Fatalf("http_get delayed failed: %v", err)
	}
	dur := time.Since(start)
	if body, _ := outH["body"].(string); body != "late" {
		t.Fatalf("unexpected body delayed: %q", body)
	}
	if dur < 0 || dur > 5*time.Second {
		t.Fatalf("unexpected duration: %v", dur)
	}
}
