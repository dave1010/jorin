package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestToolsManifestAndRegistryKeys(t *testing.T) {
	m := toolsManifest()
	if len(m) == 0 {
		t.Fatalf("expected non-empty tools manifest")
	}
	// ensure expected function names present
	expects := map[string]bool{"shell": false, "read_file": false, "write_file": false, "http_get": false}
	for _, it := range m {
		if it.Type != "function" {
			continue
		}
		expects[it.Function.Name] = true
	}
	for k, seen := range expects {
		if !seen {
			t.Fatalf("expected tool %q in manifest", k)
		}
	}

	r := registry()
	for k := range expects {
		if _, ok := r[k]; !ok {
			t.Fatalf("expected registry to contain %q", k)
		}
	}
}

func TestReadFileLargeAndWriteFileBytes(t *testing.T) {
	r := registry()
	// create large file >200k
	tmp := t.TempDir()
	f := filepath.Join(tmp, "big.txt")
	b := make([]byte, 210000)
	for i := range b {
		b[i] = 'A' + byte(i%26)
	}
	if err := os.WriteFile(f, b, 0o644); err != nil {
		t.Fatalf("write big file: %v", err)
	}
	out, err := r["read_file"](map[string]any{"path": f}, &Policy{})
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}
	text, _ := out["text"].(string)
	trunc, _ := out["truncated"].(bool)
	if !trunc {
		t.Fatalf("expected truncated=true for large file")
	}
	if len(text) != 200000 {
		t.Fatalf("expected text length 200000, got %d", len(text))
	}

	// write_file should report bytes and create file
	wf := r["write_file"]
	d := filepath.Join(tmp, "sub")
	p := filepath.Join(d, "out.txt")
	outw, err := wf(map[string]any{"path": p, "text": "hello bytes"}, &Policy{})
	if err != nil {
		t.Fatalf("write_file failed: %v", err)
	}
	if ok, _ := outw["ok"].(bool); !ok {
		t.Fatalf("expected ok true from write_file, got %#v", outw)
	}
	if bs, _ := outw["bytes"].(int); bs != len("hello bytes") {
		t.Fatalf("expected bytes=%d got %#v", len("hello bytes"), outw["bytes"])
	}
	// file exists
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected file created: %v", err)
	}
}

func TestHttpGetTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	}))
	defer srv.Close()

	r := registry()
	out, err := r["http_get"](map[string]any{"url": srv.URL}, &Policy{})
	if err != nil {
		t.Fatalf("http_get failed: %v", err)
	}
	if status, _ := out["status"].(int); status != 200 {
		// sometimes numbers decode as float64 when coming from JSON paths - but here direct
		// just check via type switch
		// accept float64 too
		if f, ok := out["status"].(float64); !ok || int(f) != 200 {
			t.Fatalf("unexpected status: %#v", out["status"])
		}
	}
	if body, _ := out["body"].(string); body != "pong" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestShellPolicyAllowDenyDryAndCWD(t *testing.T) {
	r := registry()
	shell := r["shell"]

	// missing cmd
	if _, err := shell(map[string]any{}, &Policy{}); err == nil {
		t.Fatalf("expected error for missing cmd")
	}

	// deny should block
	out, _ := shell(map[string]any{"cmd": "do forbidden stuff"}, &Policy{Deny: []string{"forbidden"}})
	if out["error"] != "denied by policy" {
		t.Fatalf("expected denied by policy, got %#v", out)
	}

	// allow list present requires allowed substring
	out, _ = shell(map[string]any{"cmd": "run ALLOW_ME now"}, &Policy{Allow: []string{"ALLOW_ME"}})
	if _, ok := out["dry_run"]; ok {
		// dry_run not set; ok
	}
	out, _ = shell(map[string]any{"cmd": "nope"}, &Policy{Allow: []string{"ALLOW_ME"}})
	if out["error"] != "not allowed by policy" {
		t.Fatalf("expected not allowed by policy, got %#v", out)
	}

	// dry run
	out, _ = shell(map[string]any{"cmd": "echo hi"}, &Policy{DryShell: true})
	if dr, _ := out["dry_run"].(bool); !dr {
		t.Fatalf("expected dry_run true, got %#v", out)
	}
	if cmd, _ := out["cmd"].(string); cmd != "echo hi" {
		t.Fatalf("unexpected cmd in dry run: %v", cmd)
	}

	// actual execution: echo and exit code
	out, err := shell(map[string]any{"cmd": "echo -n DONE; exit 0"}, &Policy{})
	if err != nil {
		t.Fatalf("shell execution failed: %v", err)
	}
	if rc, ok := out["returncode"].(int); ok {
		if rc != 0 {
			t.Fatalf("expected rc 0, got %d", rc)
		}
	} else if f, ok := out["returncode"].(float64); ok {
		if int(f) != 0 {
			t.Fatalf("expected rc 0, got %v", out["returncode"])
		}
	} else {
		t.Fatalf("unexpected returncode type: %#v", out["returncode"])
	}
	if s, _ := out["stdout"].(string); s != "DONE" {
		t.Fatalf("unexpected stdout: %q", s)
	}

	// test CWD influence: create script that prints pwd to a file
	tmp := t.TempDir()
	script := filepath.Join(tmp, "writecwd.sh")
	sh := "#!/bin/sh\necho -n $(pwd) > outpwd.txt"
	if err := os.WriteFile(script, []byte(sh), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	// run from tmp dir
	out, err = shell(map[string]any{"cmd": "./writecwd.sh"}, &Policy{CWD: tmp})
	if err != nil {
		t.Fatalf("shell CWD exec failed: %v", err)
	}
	// read out file
	b, err := os.ReadFile(filepath.Join(tmp, "outpwd.txt"))
	if err != nil {
		t.Fatalf("read outpwd failed: %v", err)
	}
	if string(b) != tmp {
		t.Fatalf("expected outpwd to be %q got %q", tmp, string(b))
	}
}

func TestHttpGetTimeout(t *testing.T) {
	// server that delays
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte("late"))
	}))
	defer srv.Close()

	r := registry()
	start := time.Now()
	out, err := r["http_get"](map[string]any{"url": srv.URL}, &Policy{})
	dur := time.Since(start)
	if err != nil {
		t.Fatalf("http_get delayed failed: %v", err)
	}
	// should have succeeded and not timed out (timeout is 15s)
	if out["body"] != "late" {
		t.Fatalf("unexpected body: %#v", out["body"])
	}
	if dur < 0 || dur > 5*time.Second {
		t.Fatalf("unexpected duration: %v", dur)
	}
}
