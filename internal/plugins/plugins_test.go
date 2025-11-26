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
	metadata = map[string]struct {
		Desc string
		Sub  map[string]string
	}{}
	modelProvider = nil
	mu.Unlock()

	h := func(ctx context.Context, name string, args []string, raw string, out io.Writer, errOut io.Writer) (bool, error) {
		_, _ = out.Write([]byte("ok"))
		return true, nil
	}
	p := &Plugin{Name: "p1", Description: "desc", Commands: map[string]CommandDef{"c1": {Description: "cd1", Handler: h, Subcommands: map[string]CommandDef{"sub": {Description: "sdesc", Handler: h}}}}}
	RegisterPlugin(p)

	list := ListPlugins()
	if len(list) != 1 || list[0].Name != "p1" {
		t.Fatalf("unexpected plugin list: %#v", list)
	}

	if h2, ok := LookupCommand("c1"); !ok || h2 == nil {
		t.Fatalf("expected command handler for c1")
	} else {
		var sb strings.Builder
		handled, err := h2(context.Background(), "c1", nil, "/c1", &sb, &sb)
		if err != nil || !handled || sb.String() != "ok" {
			t.Fatalf("handler invocation failed: %v %v %q", err, handled, sb.String())
		}
	}

	if h3, ok := LookupCommand("c1 sub"); !ok || h3 == nil {
		t.Fatalf("expected subcommand handler for 'c1 sub'")
	} else {
		var sb strings.Builder
		handled, err := h3(context.Background(), "c1 sub", nil, "/c1 sub", &sb, &sb)
		if err != nil || !handled || sb.String() != "ok" {
			t.Fatalf("sub handler invocation failed: %v %v %q", err, handled, sb.String())
		}
	}

	all := ListAllCommands()
	if d, ok := all["c1"]; !ok || d != "cd1" {
		t.Fatalf("ListAllCommands missing/incorrect: %#v", all)
	}

	desc, subs, ok := HelpForCommand("c1")
	if !ok || desc != "cd1" {
		t.Fatalf("HelpForCommand returned bad desc: %q", desc)
	}
	if s, ok := subs["sub"]; !ok || s != "sdesc" {
		t.Fatalf("HelpForCommand returned bad subs: %#v", subs)
	}
}

func TestModelProvider(t *testing.T) {
	// reset
	mu.Lock()
	plugins = nil
	commandMap = map[string]CommandHandler{}
	metadata = map[string]struct {
		Desc string
		Sub  map[string]string
	}{}
	modelProvider = nil
	mu.Unlock()

	SetModelProvider(func() string { return "mymodel" })
	if got := Model(); got != "mymodel" {
		t.Fatalf("unexpected model: %q", got)
	}
}
