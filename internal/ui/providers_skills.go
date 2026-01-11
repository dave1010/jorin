package ui

import (
	"os"
	"path/filepath"
	"strings"
)

type skillMetadata struct {
	name        string
	description string
}

// skillsProvider appends skill descriptions from ~/.jorin/skills.
type skillsProvider struct{}

func (skillsProvider) Provide() string {
	skillsDir, err := skillsDirPath()
	if err != nil {
		return ""
	}
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return ""
	}
	var skills []skillMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		content, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}
		name, desc := parseSkillFrontmatter(string(content))
		if name == "" {
			name = entry.Name()
		}
		if desc == "" {
			continue
		}
		skills = append(skills, skillMetadata{name: name, description: desc})
	}
	if len(skills) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Skills\nYou have new skills. If any skill might be relevant then you MUST read `~/.jorin/skills/<skill-name>/SKILL.md`. Skills available:\n")
	for _, skill := range skills {
		b.WriteString("- ")
		b.WriteString(skill.name)
		b.WriteString(": ")
		b.WriteString(skill.description)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func skillsDirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jorin", "skills"), nil
}

func parseSkillFrontmatter(content string) (string, string) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", ""
	}
	var name string
	var description string
	for _, line := range lines[1:] {
		trim := strings.TrimSpace(line)
		if trim == "---" {
			break
		}
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		parts := strings.SplitN(trim, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		switch key {
		case "name":
			name = value
		case "description":
			description = value
		}
	}
	return name, description
}

func init() {
	RegisterPromptProvider(skillsProvider{})
}
