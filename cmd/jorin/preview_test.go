package main

import "testing"

func TestPreview(t *testing.T) {
	s := "line1\nline2\nline3"

	if got := preview(s, 100); got == "" {
		t.Fatalf("preview returned empty")
	}
	if got := preview(s, 100); got == "line1\nline2\nline3" {
		t.Fatalf("preview did not replace newlines")
	}

	long := "aaaaaaaaaa"
	if got := preview(long, 5); got != "aaaaa..." {
		t.Fatalf("preview truncation mismatch: %s", got)
	}
}
