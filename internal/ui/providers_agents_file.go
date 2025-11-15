package ui

import "os"

// agentsFileProvider appends AGENTS.md contents when present.
type agentsFileProvider struct{}

func (agentsFileProvider) Provide() string {
	if _, err := os.Stat("AGENTS.md"); err == nil {
		if b, err := os.ReadFile("AGENTS.md"); err == nil {
			return "Project-specific instructions from AGENTS.md:\n<agents>\n" + string(b) + "</agents>\n"
		}
	}
	return ""
}

func init() {
	RegisterPromptProvider(agentsFileProvider{})
}
