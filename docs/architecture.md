# Architecture overview

This project is intentionally small and structured for auditability.

Main components

- cmd/jorin: CLI — flag parsing, REPL orchestration
- internal/agent: constructs prompts, conversation state, and tool orchestration
- internal/openai: client wrapper for OpenAI-compatible APIs
- internal/tools: tool implementations (shell, read_file, write_file, http_get)

Tools are registered in a registry and invoked via a narrow interface that
receives parameters and returns structured results. Policy checks (readonly,
allow/deny, dry-shell) are evaluated before invoking tools with side effects.

This separation keeps network and filesystem access points concentrated and
easy to review.

## System prompt extensibility

The system prompt that is sent to the LLM is built from modular "prompt
providers". Providers implement a simple interface and can be registered to
contribute parts of the overall system prompt. This makes it easy to add
project-specific instructions, runtime context, or plugin-provided guidance
without modifying core code.

- Providers implement ui.PromptProvider (Provide() string) and are registered
  in init() functions.
- Providers are concatenated in registration order with blank lines between
  sections.
- Default providers include the immutable base instructions, AGENTS.md (when
  present), personal Skills from ~/.jorin/skills (SKILL.md descriptions), and
  Situations from ~/.jorin/situations and ./.jorin/situations (executables that
  emit contextual snippets wrapped in XML-like tags). The repository ships
  runtime context Situations in .jorin/situations for git status, OS uname, and
  tools on PATH.

This is designed for future plugin integration: a plugin can register a
provider during its init/startup to contribute to the system prompt.
