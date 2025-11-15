package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/types"
	"github.com/dave1010/jorin/internal/ui/commands"
)

// minimal mock agent that echoes last user message
type mockAgent struct{}

func (m *mockAgent) ChatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	if len(msgs) == 0 {
		return msgs, "", nil
	}
	last := msgs[len(msgs)-1]
	return msgs, "ECHO: " + last.Content, nil
}

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
	var a agent.Agent = &mockAgent{}
	if err := StartREPL(ctx, a, "test-model", pol, in, out, errOut, cfg, h, hist); err != nil {
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
	// also ensure output contains echo
	if !strings.Contains(out.String(), "ECHO: hello") {
		t.Fatalf("expected model echo in out: %s", out.String())
	}
}
