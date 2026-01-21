package main

import "testing"

func TestTail(t *testing.T) {
	s := "abcdefghijklmnopqrstuvwxyz"

	if got := tail(s, 5); got != "vwxyz" {
		t.Fatalf("tail mismatch: %s", got)
	}
	if got := tail(s, 100); got != s {
		t.Fatalf("tail should return whole string when n>=len: %s", got)
	}
}
