# Security and tool permissions

jorin exposes a small set of tools to the agent; this document explains the
runtime controls and recommended safe configurations.

Tools available to the agent:

- shell: execute shell commands (subject to allow/deny/dry-run)
- read_file: read files
- write_file: write files (can be disabled with --readonly)
- http_get: unauthenticated HTTP GET requests

Runtime policy controls

- --readonly: disable write_file calls
- --dry-shell: prevent actual shell execution; commands are reported only
- --allow: one or more allowlist substrings; a shell command must match at
  least one to be executed
- --deny: one or more denylist substrings; any match blocks execution
- --cwd: working directory for tool calls

Guidance

- For untrusted environments, prefer `--readonly --dry-shell` and tight
  `--allow`/`--deny` lists.
- Use repository-level AGENTS.md to provide project-specific constraints and
  examples. The CLI appends AGENTS.md contents to the system prompt when
  present in the working directory.
