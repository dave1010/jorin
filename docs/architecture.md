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
