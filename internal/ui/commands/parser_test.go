package commands

import (
	"testing"
)

func TestParseSimple(t *testing.T) {
	c, err := Parse("/help", "/", "\\")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "help" {
		t.Fatalf("expected help, got %s", c.Name)
	}
}

func TestParseArgs(t *testing.T) {
	c, err := Parse(`/run "arg with spaces" 'single'`, "/", "\\")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "run" {
		t.Fatalf("expected run, got %s", c.Name)
	}
	if len(c.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(c.Args))
	}
}

func TestEscape(t *testing.T) {
	// a single backslash before slash: "\/help"
	c, err := Parse(`\/help`, "/", "\\")
	if err == nil {
		t.Fatalf("expected escaped error, got nil; raw=%s", c.Raw)
	}
	if c.Raw != "/help" {
		t.Fatalf("unexpected raw: %s", c.Raw)
	}
}

func TestNotCommand(t *testing.T) {
	_, err := Parse("nope", "/", "\\")
	if err == nil {
		t.Fatalf("expected not a command")
	}
}
