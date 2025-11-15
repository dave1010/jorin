package shell

import (
	"bytes"
	"os/exec"
)

// Runner executes shell commands. Implementations should aim to avoid
// using a full shell when not required, but the default runner uses
// "bash -lc" to preserve existing behavior.
type Runner interface {
	Run(cmd string, cwd string) (stdout string, stderr string, returncode int)
}

// DefaultRunner is used by packages that need to execute shell commands.
// Tests can replace DefaultRunner with a mock.
var DefaultRunner Runner = &LocalRunner{}

// LocalRunner executes commands using bash -lc and captures stdout/stderr.
type LocalRunner struct{}

func (l *LocalRunner) Run(cmd string, cwd string) (string, string, int) {
	c := exec.Command("bash", "-lc", cmd)
	if cwd != "" {
		c.Dir = cwd
	}
	var out bytes.Buffer
	var errb bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errb
	cErr := c.Run()
	rc := 0
	if cErr != nil {
		if ee, ok := cErr.(*exec.ExitError); ok {
			rc = ee.ExitCode()
		} else {
			rc = 1
		}
	}
	return out.String(), errb.String(), rc
}
