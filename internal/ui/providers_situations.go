package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type situationMetadata struct {
	name        string
	description string
	run         string
}

type situationEntry struct {
	dir      string
	metadata situationMetadata
}

// situationsProvider appends situation output from ~/.jorin/situations and ./.jorin/situations.
type situationsProvider struct{}

func (situationsProvider) Provide() string {
	entries := situationsEntries()
	if len(entries) == 0 {
		return ""
	}
	var outputs []string
	for _, entry := range entries {
		name := entry.metadata.name
		if name == "" {
			name = filepath.Base(entry.dir)
		}
		run := entry.metadata.run
		if run == "" {
			continue
		}
		runPath := filepath.Join(entry.dir, run)
		out, err := execSituation(runPath)
		if err != nil {
			continue
		}
		trimmed := strings.TrimSpace(out)
		if trimmed == "" {
			continue
		}
		outputs = append(outputs, "<"+name+">\n"+trimmed+"\n</"+name+">")
	}
	if len(outputs) == 0 {
		return ""
	}
	return "## Situations\n" + strings.Join(outputs, "\n")
}

func situationsEntries() []situationEntry {
	entries := []situationEntry{}
	for _, dir := range situationsDirPaths() {
		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range dirEntries {
			if !entry.IsDir() {
				continue
			}
			situationDir := filepath.Join(dir, entry.Name())
			metaPath := filepath.Join(situationDir, "SITUATION.yaml")
			content, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}
			meta := parseSituationMetadata(string(content))
			if meta.run == "" {
				continue
			}
			entries = append(entries, situationEntry{
				dir:      situationDir,
				metadata: meta,
			})
		}
	}
	return entries
}

func situationsDirPaths() []string {
	paths := []string{}
	if wd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(wd, ".jorin", "situations"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".jorin", "situations"))
	}
	return paths
}

func parseSituationMetadata(content string) situationMetadata {
	lines := strings.Split(content, "\n")
	meta := situationMetadata{}
	for _, line := range lines {
		trim := strings.TrimSpace(line)
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
			meta.name = value
		case "description":
			meta.description = value
		case "run":
			meta.run = value
		}
	}
	return meta
}

func execSituation(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(path)
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), "JORIN_PWD="+wd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func init() {
	RegisterPromptProvider(situationsProvider{})
}
