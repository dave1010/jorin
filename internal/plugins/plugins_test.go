package plugins

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestRegisterAndListPlugins(t *testing.T) {
	// reset global state for test
	mu.Lock()
	plugins = nil
	commandMap = map[string]CommandHandler{}
	modelProvider = nil
	mu.Unlock()

	p := &Plugin{Name: "p1", Description: "desc", Commands: map[string]CommandHandler{"c1": func(ctx context.Context, name string, args []string, raw string, out io.Writer, errOut io.Writer) (bool, error) {
		_, _ = out.Write([]byte("ok"))
		return true, nil
	}}}
	RegisterPlugin(p)

	list := ListPlugins()
	if len(list) != 1 || list[0].Name != "p1" {
		t.Fatalf("unexpected plugin list: %#v", list)
	}

	if h, ok := LookupCommand("c1"); !ok || h == nil {
		t.Fatalf("expected command handler for c1")
	} else {
		var sb strings.Builder
		handled, err := h(context.Background(), "c1", nil, "/c1", &sb, &sb)
		if err != nil || !handled || sb.String() != "ok" {
			t.Fatalf("handler invocation failed: %v %v %q", err, handled, sb.String())
		}
	}
}

func TestModelProvider(t *testing.T) {
	// reset
	mu.Lock()
	plugins = nil
	commandMap = map[string]CommandHandler{}
	modelProvider = nil
	mu.Unlock()

	SetModelProvider(func() string { return "mymodel" })
	if got := Model(); got != "mymodel" {
		t.Fatalf("unexpected model: %q", got)
	}
}
