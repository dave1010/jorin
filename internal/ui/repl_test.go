package ui

import (
	"bytes"
	"context"
	"testing"

	"github.com/dave1010/jorin/internal/types"
	"github.com/dave1010/jorin/internal/ui/commands"
)

// minimal mock agent is not wired; tests focus on parser and handler wiring

func TestREPLHandlerDispatch(t *testing.T) {
	in := bytes.NewBufferString("/help\nhello\n\\/help\n")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	pol := &types.Policy{}
	cfg := DefaultConfig()
	// create handler with history
	hist := NewMemHistory(10)
	h := commands.NewDefaultHandler(out, errOut, hist)
	ctx := context.Background()
	// run StartREPL in the same goroutine; it will exit on EOF
	if err := StartREPL(ctx, "test-model", pol, in, out, errOut, cfg, h, hist); err != nil {
		t.Fatalf("StartREPL failed: %v", err)
	}
	// ensure history recorded the non-command "hello"
	list := hist.List(10)
	found := false
	for _, l := range list {
		if l == "hello" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected history to contain 'hello', got %v", list)
	}
}
