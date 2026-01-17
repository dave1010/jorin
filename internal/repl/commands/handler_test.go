package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/dave1010/jorin/internal/plugins"
)

func TestHelpIncludesPluginCommands(t *testing.T) {
	// reset plugin registry
	plugins.SetModelProvider(nil)
	// since plugins keeps package-level state, we reset internals via
	// plugin package's unexported vars using RegisterPlugin behavior. For
	// tests, create a fresh plugin with a command and rely on listing.

	p := &plugins.Plugin{
		Name:        "ptest",
		Description: "test plugin",
		Commands: map[string]plugins.CommandDef{
			"p": {Description: "top p", Subcommands: map[string]plugins.CommandDef{"sub": {Description: "sub p"}}},
		},
	}
	plugins.RegisterPlugin(p)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	hist := &memHistory{lines: []string{}}
	h := NewDefaultHandler(out, errOut, hist, nil)
	ctx := context.Background()
	// ask for /help (no topic) to include plugin commands
	if ok, err := h.Handle(ctx, Command{Name: "help", Args: nil, Raw: "/help"}); !ok || err != nil {
		t.Fatalf("help command failed: %v %v", ok, err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Plugin commands:")) {
		t.Fatalf("expected plugin commands in help output: %s", out.String())
	}
	// ask for /help p (topic = plugin command)
	out.Reset()
	if ok, err := h.Handle(ctx, Command{Name: "help", Args: []string{"p"}, Raw: "/help p"}); !ok || err != nil {
		t.Fatalf("help specific failed: %v %v", ok, err)
	}
	if !bytes.Contains(out.Bytes(), []byte("top p")) {
		t.Fatalf("expected top p description in help output: %s", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Subcommands:")) {
		t.Fatalf("expected subcommands section in help output: %s", out.String())
	}
}

// memHistory implements History for tests
type memHistory struct {
	lines []string
}

func (m *memHistory) Add(line string)         { m.lines = append(m.lines, line) }
func (m *memHistory) List(limit int) []string { return m.lines }
