package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadScriptPrompt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review-code.jorin")
	content := "#!/usr/bin/env jorin\nEnsure SOLID principles are followed.\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	prompt, ok, err := loadScriptPrompt(path)
	if err != nil {
		t.Fatalf("loadScriptPrompt failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected script detection to be true")
	}
	if prompt != "Ensure SOLID principles are followed.\n" {
		t.Fatalf("unexpected prompt: %q", prompt)
	}
}

func TestLoadScriptPromptIgnoresNonShebang(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(path, []byte("just text"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, ok, err := loadScriptPrompt(path)
	if err != nil {
		t.Fatalf("loadScriptPrompt failed: %v", err)
	}
	if ok {
		t.Fatalf("expected non-shebang file to be ignored")
	}
}

func TestResolvePromptPrefersScript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "script.jorin")
	if err := os.WriteFile(path, []byte("#!/usr/bin/env jorin\nRun checks."), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	prompt, args, err := resolvePrompt([]string{path, "alpha", "beta"})
	if err != nil {
		t.Fatalf("resolvePrompt failed: %v", err)
	}
	if prompt != "Run checks." {
		t.Fatalf("unexpected prompt: %q", prompt)
	}
	if len(args) != 2 || args[0] != "alpha" || args[1] != "beta" {
		t.Fatalf("unexpected args: %#v", args)
	}
}
