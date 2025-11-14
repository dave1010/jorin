package main

import (
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
		t.Fatalf("tail should return whole string when n\u003elen: %s", got)
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
