package prompt

import "sync/atomic"

const ralphPrompt = `## Ralph Wiggum Loop Mode
You are operating in "Ralph Wiggum" loop mode: an iterative, self-referential development loop that feeds your output back into the next iteration until completion.

- Iteration beats perfection: make steady, incremental progress.
- Failures are data: report what failed and what to try next.
- Operator skill matters: be explicit about the exact inputs and commands to run.
- Persistence wins: keep moving the task forward each iteration.

When the task is fully complete, end your response with the exact word "DONE" on its own line.`

var ralphEnabled atomic.Bool

// EnableRalph toggles Ralph Wiggum loop instructions in the system prompt.
func EnableRalph() {
	ralphEnabled.Store(true)
}

// RalphEnabled reports whether Ralph Wiggum loop instructions are enabled.
func RalphEnabled() bool {
	return ralphEnabled.Load()
}

type ralphProvider struct{}

func (ralphProvider) Provide() string {
	if !ralphEnabled.Load() {
		return ""
	}
	return ralphPrompt
}

func init() {
	RegisterPromptProvider(ralphProvider{})
}
