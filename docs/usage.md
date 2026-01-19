# Usage

This document covers how to run and interact with jorin from the command line.

## Quick start

Show help:

```bash
jorin --help
```

Start the REPL (default when invoked with no args):

```bash
jorin
```

Send a single prompt from the command line or a script:

```bash
jorin "Refactor function X to be smaller"
# or
echo "Add logging to foo()" | jorin
```

## Configuration

Set these environment variables before running jorin:

| Variable | Purpose |
| --- | --- |
| `OPENAI_API_KEY` | API key for OpenAI-compatible endpoints. Required. |
| `OPENAI_BASE_URL` | Overrides the API base URL (default: `https://api.openai.com`). |
| `NO_COLOR` | Disables ANSI color output when set. |
| `TERM` | If set to `dumb`, disables color output. |

## CLI reference

### Invocation modes

- **REPL (interactive)**: Start with no args or `--repl`.
- **Single prompt**: Provide a quoted prompt or pipe stdin. Remaining CLI args
  are joined into the prompt string.
- **Prompt file (auto)**: If the first arg is a readable file (not a directory),
  the file contents become the prompt and remaining args are appended as
  arguments. Use `--prompt` to disable auto file loading or `--prompt-file` to
  require it.

Examples:

```bash
jorin --repl
jorin "Summarize the last test run"
echo "Add logging to foo()" | jorin
./review-code.jorin --target src/
jorin --prompt "review-code.jorin --target src/"
jorin --prompt-file prompts/review-code.jorin --target src/
```

### Prompt files, scripts, and stdin

Jorin can load prompt files directly. If the file starts with a `jorin`
shebang, the shebang line is ignored:

```bash
#!/usr/bin/env jorin
Ensure SOLID principles are followed.
```

When you run a prompt file, remaining arguments are appended to the prompt as
arguments, and piped stdin is appended as stdin context:

```bash
./review-code.jorin --target src/ < notes.txt
jorin prompts/review-code.jorin --target src/ < notes.txt
```

To use stdin directly with a one-off prompt:

```bash
cat document.md | jorin "Summarize the text"
```

If you omit the prompt entirely, piped stdin becomes the prompt:

```bash
cat document.md | jorin
```

### Command-line flags

| Flag | Default | Description |
| --- | --- | --- |
| `--model` | `gpt-5-mini` | Model ID sent to the API. |
| `--repl` | `false` | Start an interactive REPL. |
| `--readonly` | `false` | Disallow `write_file` tool calls. |
| `--dry-shell` | `false` | Do not execute shell commands (report them only). |
| `--allow` | (none) | Allowlist substring for shell commands. Repeatable. |
| `--deny` | (none) | Denylist substring for shell commands. Repeatable. |
| `--cwd` | (empty) | Working directory for shell tool execution. |
| `--prompt` | `false` | Treat the first argument as literal prompt text (disables prompt-file detection). |
| `--prompt-file` | `false` | Treat the first argument as a prompt file (error if not a readable file). |
| `--version` | `false` | Print version and exit. |

Notes:

- If `--allow` is provided, every shell command must match at least one
  allowlisted substring.
- If `--deny` is provided, any substring match blocks execution.
- `--cwd` applies to the `shell` tool only; read/write paths are used as given.

## REPL commands

Built-in commands:

- `/help` or `/help <topic>`: Show available commands and help topics.
- `/history [n]`: List the last `n` prompts (or all stored history).
- `/debug`: Print the full system prompt (including AGENTS.md content, Skill
  descriptions, and Situation output).

Plugin-provided commands:

- `/plugins`: List compiled-in plugins.
- `/model`: Show the currently configured model.

Plugin commands are only available when their plugin is compiled into the
binary.

## Examples

Dry-run shell mode (agent reports shell commands but does not execute them):

```bash
jorin --dry-shell "Run the tests"
```

Prevent file writes (audit mode):

```bash
jorin --readonly "Make a small change to main.go"
```

Allow/deny list examples:

```bash
# Only allow shell commands containing the substring ALLOW_ME
jorin --allow ALLOW_ME "Run the deployment script"

# Deny commands containing dangerous substrings
jorin --deny "rm -rf" --deny "passwd" "Audit the machine"
```

## Tool behavior

The agent can invoke the following tools. Each tool returns structured JSON to
the model (and a concise preview is written to stderr in the CLI).

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

## Skills and Situations

Jorin supports two prompt-context conventions: Skills (Anthropic) and
Situations (Jorin-specific).

### Skills (Anthropic convention)

Skills live under `~/.jorin/skills` or `./.jorin/skills`, one directory per
skill. Each skill directory must include a `SKILL.md` with YAML frontmatter
(`name`, `description`).

- The `description` is required; skills without a description are skipped.
- The `name` defaults to the directory name when omitted.
- Jorin injects the skill descriptions into the system prompt and instructs the
  agent to read the full `SKILL.md` when a skill is relevant.

Reference: [Claude Code Skills](https://code.claude.com/docs/en/skills)

### Situations (Jorin convention)

Situations are executable context providers that emit prompt snippets. Create
them under `~/.jorin/situations` or `./.jorin/situations` (project-specific).
Each situation lives in its own folder with a `SITUATION.yaml` metadata file
and an executable referenced by the `run` field.

- The `run` field is required; situations without it are ignored.
- The executable runs from the current working directory and receives
  `JORIN_PWD` pointing at that directory.
- Output is wrapped in `<name>...</name>` tags and appended to the system
  prompt. `name` defaults to the directory name when omitted.

The repository ships built-in situations under `./.jorin/situations` for
reporting git status, runtime environment, available executables, and Go module
detection.

Reference: [Agent Situations repository](https://github.com/dave1010/agent-situations)
Blog: [Giving coding agents situational awareness (from shell prompts to agent prompts)](https://dave.engineer/blog/2026/01/agent-situations/)

Example:

```text
~/.jorin/situations/php/SITUATION.yaml
name: php
description: Detect PHP projects via .php-version.
run: run
```

```bash
~/.jorin/situations/php/run
#!/usr/bin/env bash
set -euo pipefail

if [[ -f ".php-version" ]]; then
  echo "This is a PHP project, requiring version $(cat .php-version) at minimum."
fi
```

## Plugin system

Jorin supports compiled-in plugins that can register additional slash commands
(including nested subcommands). Plugins are compiled into the binary and
register themselves at init().

Available plugin features:

- Register top-level commands with a description and handler.
- Register subcommands under a top-level command; subcommands also have their
  own descriptions and handlers.
- Plugins can obtain runtime context like the current model by the host setting
  a model provider callback.

Provided built-in plugin:

- model-plugin
  - /plugins — lists compiled-in plugins (name and description)
  - /model — prints the currently selected model (reads from the host-provided
    model provider)

How to write and register a plugin (compiled-in)

- Create a package under internal/plugins or another package that imports
  github.com/dave1010/jorin/internal/plugins.
- Create a Plugin value and call plugins.RegisterPlugin in an init() function.
- Provide CommandDef entries with Description, Handler, and optional
  Subcommands.

Example (informal):

```go
func init() {
  p := &plugins.Plugin{
    Name: "my-plugin",
    Description: "Adds /thing commands",
    Commands: map[string]plugins.CommandDef{
      "thing": {
        Description: "manage things",
        Handler: myThingHandler,
        Subcommands: map[string]plugins.CommandDef{
          "list": { Description: "list things", Handler: myThingListHandler },
        },
      },
    },
  }
  plugins.RegisterPlugin(p)
}
```

Using help and plugin commands

- /help — lists builtin help topics and plugin commands.
- /help <topic> — shows builtin help for a topic (e.g. `repl`) or plugin help
  for a command with descriptions and subcommands.
- Invoke plugin commands as usual: `/command` or include a subcommand
  `/command sub`.

## Troubleshooting

### API errors

#### `API 401` or `API 403`

Likely causes:

- Missing or invalid `OPENAI_API_KEY`.
- An API key that is not authorized for the requested model.

Fix:

- Export a valid API key and retry.
- Verify your account has access to the selected model.

#### `API 429`

Likely causes:

- Rate limits or quota exhaustion from the API provider.

Fix:

- Wait and retry, or switch to a different model/account with more quota.

#### `API 5xx`

Likely causes:

- Temporary service outage.

Fix:

- Retry after a short delay, or use a different `OPENAI_BASE_URL`.

### Shell/tool issues

#### Command blocked by policy

If a tool call returns `{"error":"denied by policy"}` or
`{"error":"not allowed by policy"}`, the `--allow` and `--deny` flags are
intervening.

Fix:

- Remove or loosen `--deny` substrings.
- Add an `--allow` substring that matches the full command.

#### No file output

If a `write_file` tool call returns `{"error":"readonly session"}`, the CLI
was started with `--readonly`.

Fix:

- Remove the `--readonly` flag.

### Color or formatting issues

If you see garbled ANSI output or want plain text:

- Set `NO_COLOR=1`, or
- Use a terminal that does not report `TERM=dumb`.

## Security notes

For more details about shell policy and tool permissions, see the
[Security notes](security.md).
