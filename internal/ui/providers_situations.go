package ui

import (
	"fmt"
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
			// nothing to run; add debug block so it's visible during development
			outputs = append(outputs, "\u003c"+name+"-debug\u003e\nno run field in SITUATION.yaml\n\u003c/"+name+"-debug\u003e")
			continue
		}
		runPath := filepath.Join(entry.dir, run)
		out, err := execSituation(runPath)
		trimmed := strings.TrimSpace(out)
		if err != nil {
			// collect debug information (error and any output/stderr)
			statInfo := ""
			if fi, statErr := os.Stat(runPath); statErr == nil {
				statInfo = fmt.Sprintf("mode=%v size=%d", fi.Mode(), fi.Size())
			} else {
				statInfo = fmt.Sprintf("stat error: %v", statErr)
			}
			debug := fmt.Sprintf("\u003c%s-error\u003e\npath: %s\nerror: %v\n%s\noutput:\n%s\n\u003c/%s-error\u003e", name, runPath, err, statInfo, strings.TrimSpace(out), name)
			// if there is meaningful stdout/stderr, include it as the main output as well
			if trimmed == "" {
				outputs = append(outputs, debug)
				continue
			} else {
				outputs = append(outputs, "\u003c"+name+"\u003e\n"+trimmed+"\n\u003c/"+name+"\u003e")
				outputs = append(outputs, debug)
				continue
			}
		}
		if trimmed == "" {
			// include a debug block so empty output cases are visible during troubleshooting
			outputs = append(outputs, "\u003c"+name+"-debug\u003e\nempty output from run: "+runPath+"\n\u003c/"+name+"-debug\u003e")
			continue
		}
		outputs = append(outputs, "\u003c"+name+"\u003e\n"+trimmed+"\n\u003c/"+name+"\u003e")
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

	// Try executing the path directly first. On some platforms (eg Android/Termux)
	// the interpreter referenced by a shebang (e.g. /usr/bin/env) may not exist
	// at that exact path which causes exec to fail with "no such file or directory".
	// In that case fall back to running the script via common shells.
	cmd := exec.Command(path)
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), "JORIN_PWD="+wd)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), nil
	}

	// Attempt common fallbacks: bash then sh
	cmd = exec.Command("bash", path)
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), "JORIN_PWD="+wd)
	out2, err2 := cmd.CombinedOutput()
	if err2 == nil {
		return string(out2), nil
	}

	cmd = exec.Command("sh", path)
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), "JORIN_PWD="+wd)
	out3, err3 := cmd.CombinedOutput()
	if err3 == nil {
		return string(out3), nil
	}

	// If none succeeded, return the outputs we have and the original error.
	combined := strings.TrimSpace(string(out) + "\n" + string(out2) + "\n" + string(out3))
	return combined, err
}

func init() {
	RegisterPromptProvider(situationsProvider{})
}
