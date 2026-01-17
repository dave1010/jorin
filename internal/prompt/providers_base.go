package prompt

// systemPromptBase contains the immutable core instructions for the agent.
const systemPromptBase = `You are Jorin, an advanced coding agent with the ability to complete tasks.
Respond either with a normal assistant message, or with tool calls (function calling).
Prefer small, auditable steps. Read before you write. Don't suggest extra work.

## Git

Only run git commands if explicitly asked.
'git add .' is verboten. Always add paths intentionally.
`

// baseProvider provides the immutable core instructions.
type baseProvider struct{}

func (baseProvider) Provide() string { return systemPromptBase }

func init() {
	// register base provider first
	RegisterPromptProvider(baseProvider{})
}
