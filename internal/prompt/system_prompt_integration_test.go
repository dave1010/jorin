package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSystemPromptIncludesProviders(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", tmp)

	agentsContent := "Follow the test instructions."
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte(agentsContent), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	skillDir := filepath.Join(tmp, ".jorin", "skills", "demo-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	skillContent := strings.Join([]string{
		"---",
		"name: Demo Skill",
		"description: helps with demo testing",
		"---",
		"",
		"Extra skill details.",
	}, "\n")
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	situationDir := filepath.Join(tmp, ".jorin", "situations", "demo")
	if err := os.MkdirAll(situationDir, 0o755); err != nil {
		t.Fatalf("mkdir situation dir: %v", err)
	}
	situationMeta := strings.Join([]string{
		"name: demo",
		"description: demo situation",
		"run: run.sh",
	}, "\n")
	if err := os.WriteFile(filepath.Join(situationDir, "SITUATION.yaml"), []byte(situationMeta), 0o644); err != nil {
		t.Fatalf("write SITUATION.yaml: %v", err)
	}
	runPath := filepath.Join(situationDir, "run.sh")
	if err := os.WriteFile(runPath, []byte("#!/bin/sh\necho 'situation ok'\n"), 0o755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	prompt := SystemPrompt()
	if !strings.Contains(prompt, "You are Jorin") {
		t.Errorf("expected base prompt to be included")
	}
	if !strings.Contains(prompt, agentsContent) {
		t.Errorf("expected AGENTS.md content in prompt")
	}
	if !strings.Contains(prompt, "## Skills") || !strings.Contains(prompt, "Demo Skill: helps with demo testing") {
		t.Errorf("expected skills to be listed in prompt")
	}
	if !strings.Contains(prompt, "<demo>\nsituation ok\n</demo>") {
		t.Errorf("expected situation output in prompt")
	}
}
