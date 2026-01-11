# CLI Reference

This document collects the full command-line, runtime, and tool reference for
`jorin`. It complements the usage guide by focusing on enumerations, defaults,
and behavior details.

## Invocation modes

- **REPL (interactive)**: Start with no args or `--repl`.
- **Single prompt**: Provide a quoted prompt or pipe stdin. Remaining CLI args
  are joined into the prompt string.

Examples:

```bash
jorin --repl
jorin "Summarize the last test run"
echo "Add logging to foo()" | jorin
```

## Command-line flags

| Flag | Default | Description |
| --- | --- | --- |
| `--model` | `gpt-5-mini` | Model ID sent to the API. |
| `--repl` | `false` | Start an interactive REPL. |
| `--readonly` | `false` | Disallow `write_file` tool calls. |
| `--dry-shell` | `false` | Do not execute shell commands (report them only). |
| `--allow` | (none) | Allowlist substring for shell commands. Repeatable. |
| `--deny` | (none) | Denylist substring for shell commands. Repeatable. |
| `--cwd` | (empty) | Working directory for shell tool execution. |
| `--version` | `false` | Print version and exit. |

Notes:

- If `--allow` is provided, **every** shell command must match at least one
  allowlisted substring.
- If `--deny` is provided, any substring match blocks execution.
- `--cwd` applies to the `shell` tool only; read/write paths are used as given.

## Environment variables

| Variable | Purpose |
| --- | --- |
| `OPENAI_API_KEY` | API key for OpenAI-compatible endpoints. Required. |
| `OPENAI_BASE_URL` | Overrides the API base URL (default: `https://api.openai.com`). |
| `NO_COLOR` | Disables ANSI color output when set. |
| `TERM` | If set to `dumb`, disables color output. |

## REPL commands

Built-in commands:

- `/help` or `/help <topic>`: Show available commands and help topics.
- `/history [n]`: List the last `n` prompts (or all stored history).
- `/debug`: Print the full system prompt (including AGENTS.md content and Skill descriptions).

Plugin-provided commands:

- `/plugins`: List compiled-in plugins.
- `/model`: Show the currently configured model.

Plugin commands are only available when their plugin is compiled into the
binary.

## Tool behavior

The agent can invoke the following tools. Each tool returns structured JSON
to the model (and a concise preview is written to stderr in the CLI).

### `shell`

Executes a shell command via `bash -lc`.

Response fields:

- `returncode`: integer exit status.
- `stdout`: last 8000 characters of stdout.
- `stderr`: last 8000 characters of stderr.

Policy behavior:

- `--dry-shell` returns `{ "dry_run": true, "cmd": "..." }`.
- `--allow`/`--deny` are evaluated as substring matches before execution.

### `read_file`

Reads a UTF-8 text file from disk.

Response fields:

- `text`: file contents (truncated at 200,000 characters).
- `truncated`: `true` when truncation occurs.

### `write_file`

Writes UTF-8 text to disk, creating parent directories as needed.

Response fields:

- `ok`: boolean success flag.
- `bytes`: number of bytes written.

Policy behavior:

- `--readonly` returns `{ "error": "readonly session" }` without writing.

### `http_get`

Fetches a URL with a 15-second timeout and returns up to 8000 bytes.

Response fields:

- `status`: HTTP status code.
- `body`: response body (truncated to 8000 bytes).

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success. |
| `1` | Runtime or API error. |
| `2` | No prompt provided outside REPL mode. |
