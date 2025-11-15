package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

func TestRegistryReadWriteAndShell(t *testing.T) {
	r := registry()
	// read_file missing path
	if _, err := r["read_file"](map[string]any{}, &types.Policy{}); err == nil {
		t.Fatalf("expected error for missing path")
	}

	// create temp file
	tmpDir := t.TempDir()
	fpath := filepath.Join(tmpDir, "t.txt")
	if err := os.WriteFile(fpath, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	out, err := r["read_file"](map[string]any{"path": fpath}, &types.Policy{})
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}
	if out["text"] != "hello world" {
		t.Fatalf("unexpected read_file text: %#v", out["text"])
	}

	// write_file readonly
	wf := r["write_file"]
	ro := &types.Policy{Readonly: true}
	outw, err := wf(map[string]any{"path": filepath.Join(tmpDir, "x.txt"), "text": "ok"}, ro)
	if err != nil {
		t.Fatalf("write_file returned err: %v", err)
	}
	if outw["error"] != "readonly session" {
		t.Fatalf("expected readonly error, got: %#v", outw)
	}

	// write_file success
	outw, err = wf(map[string]any{"path": filepath.Join(tmpDir, "x.txt"), "text": "ok"}, &types.Policy{})
	if err != nil {
		t.Fatalf("write_file failed: %v", err)
	}
	if outw["ok"] != true {
		t.Fatalf("expected ok true, got: %#v", outw)
	}
	if b, err := os.ReadFile(filepath.Join(tmpDir, "x.txt")); err != nil || string(b) != "ok" {
		t.Fatalf("write_file did not write content: err=%v, content=%q", err, string(b))
	}

	// shell missing cmd
	if _, err := r["shell"](map[string]any{}, &types.Policy{}); err == nil {
		t.Fatalf("expected error for missing shell cmd")
	}

	// shell deny
	polD := &types.Policy{Deny: []string{"forbidden"}}
	outS, _ := r["shell"](map[string]any{"cmd": "do something forbidden now"}, polD)
	if outS["error"] != "denied by policy" {
		t.Fatalf("expected denied by policy, got: %#v", outS)
	}

	// shell allow when allow list present
	polA := &types.Policy{Allow: []string{"ALLOW_ME"}}
	outS, _ = r["shell"](map[string]any{"cmd": "run ALLOW_ME command"}, polA)
	// not expecting a dry_run here; just ensure it returned without error
	_ = outS
	// command not allowed
	outS, _ = r["shell"](map[string]any{"cmd": "nope"}, polA)
	if outS["error"] != "not allowed by policy" {
		t.Fatalf("expected not allowed by policy, got: %#v", outS)
	}

	// dry shell
	outS, _ = r["shell"](map[string]any{"cmd": "echo hi"}, &types.Policy{DryShell: true})
	if outS["dry_run"] != true || outS["cmd"] != "echo hi" {
		t.Fatalf("dry shell unexpected: %#v", outS)
	}

	// actual shell run (simple echo)
	outS, err = r["shell"](map[string]any{"cmd": "echo hello"}, &types.Policy{})
	if err != nil {
		t.Fatalf("shell run failed: %v", err)
	}
	stdout := outS["stdout"].(string)
	if stdout == "" || !contains(stdout, "hello") {
		t.Fatalf("shell stdout missing hello: %#v", outS["stdout"])
	}
}
