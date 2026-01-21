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
		got = filepath.Clean(got)
		want = filepath.Clean(want)
		if got != want {
			t.Fatalf("dirOrDot(%q) = %q; want %q", in, got, want)
		}
	}
}
