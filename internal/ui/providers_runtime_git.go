package ui

import (
	"os"
	"strings"
)

// gitProvider reports git repository presence and path context.
type gitProvider struct{}

func (gitProvider) Provide() string {
	parts := []string{}
	if wd, err := os.Getwd(); err == nil {
		parts = append(parts, "PWD: "+wd)
	}
	if _, err := os.Stat(".git"); err == nil {
		parts = append(parts, "Git repository: yes (.git exists)")
	} else {
		parts = append(parts, "Git repository: no (.git not found)")
	}
	return strings.Join(parts, "\n")
}

func init() {
	RegisterPromptProvider(gitProvider{})
}
