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

	prompt, ok, err := loadPromptFile(path, true)
	if err != nil {
		t.Fatalf("loadPromptFile failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected script detection to be true")
	}
	if prompt != "Ensure SOLID principles are followed.\n" {
		t.Fatalf("unexpected prompt: %q", prompt)
	}
}

func TestLoadPromptFilePlainText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(path, []byte("just text\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	prompt, ok, err := loadPromptFile(path, true)
	if err != nil {
		t.Fatalf("loadPromptFile failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected plain text file to be loaded")
	}
	if prompt != "just text\n" {
		t.Fatalf("unexpected prompt: %q", prompt)
	}
}

func TestResolvePromptAutoUsesPromptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "script.jorin")
	if err := os.WriteFile(path, []byte("#!/usr/bin/env jorin\nRun checks."), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	prompt, args, err := resolvePrompt([]string{path, "alpha", "beta"}, promptModeAuto)
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

func TestResolvePromptForcedText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.txt")
	if err := os.WriteFile(path, []byte("Use the file."), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	prompt, args, err := resolvePrompt([]string{path, "extra"}, promptModeText)
	if err != nil {
		t.Fatalf("resolvePrompt failed: %v", err)
	}
	if prompt != path+" extra" {
		t.Fatalf("unexpected prompt: %q", prompt)
	}
	if args != nil {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestResolvePromptRequiresFile(t *testing.T) {
	_, _, err := resolvePrompt([]string{"missing.txt"}, promptModeFile)
	if err == nil {
		t.Fatalf("expected error for missing prompt file")
	}
}
