package ui

import (
	"os/exec"
	"strings"
)

// execsProvider reports which helpful executables are on PATH.
type execsProvider struct{}

func (execsProvider) Provide() string {
	toolsList := []string{"ag", "rg", "git", "gh", "go", "gofmt", "docker", "fzf", "python", "python3", "php", "curl", "wget"}
	found := []string{}
	for _, t := range toolsList {
		if _, err := exec.LookPath(t); err == nil {
			found = append(found, t+" ")
		}
	}
	if len(found) > 0 {
		return "Tools on PATH (others will exist too): " + strings.Join(found, ", ")
	}
	return "Tools on PATH: none of " + strings.Join(toolsList, ", ")
}

func init() {
	RegisterPromptProvider(execsProvider{})
}
