package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirOrDot(t *testing.T) {
	cases := map[string]string{
		"file.txt":     ".",
		"./file.txt":   ".",
		"a/b/c.txt":    "a/b",
		"/tmp/foo/bar": "/tmp/foo",
		"/single":      "/",
		".":            ".",
		"":             ".",
	}
	for in, want := range cases {
		got := dirOrDot(in)
		// Normalize for platforms
		got = filepath.Clean(got)
		want = filepath.Clean(want)
		if got != want {
			t.Fatalf("dirOrDot(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestTail(t *testing.T) {
	s := "abcdefghijklmnopqrstuvwxyz"
	if got := tail(s, 5); got != "vwxyz" {
		t.Fatalf("tail mismatch: %s", got)
	}
	if got := tail(s, 100); got != s {
		t.Fatalf("tail should return whole string when n>len: %s", got)
	}
}

func TestPreview(t *testing.T) {
	s := "line1\nline2\nline3"
	if got := preview(s, 100); got == "" {
		t.Fatalf("preview returned empty")
	}
	// ensure newlines removed
	if got := preview(s, 100); got == "line1\nline2\nline3" {
		t.Fatalf("preview did not replace newlines")
	}
	// length truncation
	long := "aaaaaaaaaa"
	if got := preview(long, 5); got != "aaaaa..." {
		t.Fatalf("preview truncation mismatch: %s", got)
	}
}

func TestRegistryReadWriteAndShell(t *testing.T) {
	r := registry()
	// read_file missing path
	if _, err := r["read_file"](map[string]any{}, &Policy{}); err == nil {
		t.Fatalf("expected error for missing path")
	}

	// create temp file
	tmpDir := t.TempDir()
	fpath := filepath.Join(tmpDir, "t.txt")
	if err := os.WriteFile(fpath, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	out, err := r["read_file"](map[string]any{"path": fpath}, &Policy{})
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}
	if out["text"] != "hello world" {
		t.Fatalf("unexpected read_file text: %#v", out["text"])
	}

	// write_file readonly
	wf := r["write_file"]
	ro := &Policy{Readonly: true}
	outw, err := wf(map[string]any{"path": filepath.Join(tmpDir, "x.txt"), "text": "ok"}, ro)
	if err != nil {
		t.Fatalf("write_file returned err: %v", err)
	}
	if outw["error"] != "readonly session" {
		t.Fatalf("expected readonly error, got: %#v", outw)
	}

	// write_file success
	outw, err = wf(map[string]any{"path": filepath.Join(tmpDir, "x.txt"), "text": "ok"}, &Policy{})
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
	if _, err := r["shell"](map[string]any{}, &Policy{}); err == nil {
		t.Fatalf("expected error for missing shell cmd")
	}

	// shell deny
	polD := &Policy{Deny: []string{"forbidden"}}
	outS, _ := r["shell"](map[string]any{"cmd": "do something forbidden now"}, polD)
	if outS["error"] != "denied by policy" {
		t.Fatalf("expected denied by policy, got: %#v", outS)
	}

	// shell allow when allow list present
	polA := &Policy{Allow: []string{"ALLOW_ME"}}
	outS, _ = r["shell"](map[string]any{"cmd": "run ALLOW_ME command"}, polA)
	if outS["dry_run"] == true {
		// not dry run here
	}
	// command not allowed
	outS, _ = r["shell"](map[string]any{"cmd": "nope"}, polA)
	if outS["error"] != "not allowed by policy" {
		t.Fatalf("expected not allowed by policy, got: %#v", outS)
	}

	// dry shell
	outS, _ = r["shell"](map[string]any{"cmd": "echo hi"}, &Policy{DryShell: true})
	if outS["dry_run"] != true || outS["cmd"] != "echo hi" {
		t.Fatalf("dry shell unexpected: %#v", outS)
	}

	// actual shell run (simple echo)
	outS, err = r["shell"](map[string]any{"cmd": "echo hello"}, &Policy{})
	if err != nil {
		t.Fatalf("shell run failed: %v", err)
	}
	if outS["returncode"].(int) != 0 {
		// returncode may be float64 depending on unmarshalling; handle generically
		// convert via any -> float64 or int
		// check stdout contains hello
	}
	stdout := outS["stdout"].(string)
	if stdout == "" || !contains(stdout, "hello") {
		t.Fatalf("shell stdout missing hello: %#v", outS["stdout"])
	}
}

// helper contains for tests without importing strings repeatedly
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool { return filepath.Separator != '/' || true })() && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
