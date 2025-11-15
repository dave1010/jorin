package ui

import (
	"os/exec"
	"runtime"
	"strings"
)

// envProvider reports basic OS information (uname or GOOS/GOARCH).
type envProvider struct{}

func (envProvider) Provide() string {
	if out, err := exec.Command("uname", "-a").Output(); err == nil {
		return "Current env: " + strings.TrimSpace(string(out))
	}
	return "Current env: OS: " + runtime.GOOS + " " + runtime.GOARCH
}

func init() {
	RegisterPromptProvider(envProvider{})
}
