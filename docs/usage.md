# Usage

This document covers typical user-facing ways to run and interact with jorin.

## Quick start

Build the CLI (or download a release):

```bash
make build
# or
go build -o jorin ./cmd/jorin
```

Show help:

```bash
./jorin --help
```

Start the REPL (default when invoked with no args):

```bash
./jorin --repl
# or simply
./jorin
```

Send a single prompt from the command line or a script:

```bash
./jorin "Refactor function X to be smaller"
# or
echo "Add logging to foo()" | ./jorin
```

## Key runtime flags

- --model: Model ID (default: gpt-5-mini)
- --repl: Start an interactive REPL
- --readonly: Disallow write_file operations
- --dry-shell: Do not execute shell commands (shell calls reported as dry run)
- --allow (repeatable): Allowlist substring for shell commands
- --deny (repeatable): Denylist substring for shell commands
- --cwd: Working directory for tools

Use --readonly and --dry-shell together when running in untrusted or shared
environments.

## REPL details

- Interactive terminal (tty): full line editing (cursor movement, history,
  Ctrl-C to abort input).
- Non-interactive (piped): falls back to a simple scanner mode suitable for
  deterministic scripting and tests.
- Slash commands: supported with prefix `/` (e.g. `/help`, `/history`). Escape
  the prefix with a backslash to send a literal slash (e.g. `\/help`).
- Legacy `!` prefix: lines starting with `!` are treated as shell commands.
- /debug: prints the full system prompt (including AGENTS.md content).

## Plugin system

jorin supports compiled-in plugins that can register additional slash
commands (including nested subcommands). Plugins are compiled into the binary
and register themselves at init().

Available plugin features:

- Register top-level commands with a description and handler.
- Register subcommands under a top-level command; subcommands also have their
  own descriptions and handlers.
- Plugins can obtain runtime context like the current model by the host
  setting a model provider callback.

Provided built-in plugin:

- model-plugin
  - /plugins — lists compiled-in plugins (name and description)
  - /model — prints the currently selected model (reads from the host-provided model provider)

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
- Invoke plugin commands as usual: `/command` or include a subcommand `/command sub`.

## Examples

Dry-run shell mode (agent reports shell commands but does not execute them):

```bash
./jorin --dry-shell "Run the tests"
```

Prevent file writes (audit mode):

```bash
./jorin --readonly "Make a small change to main.go"
```

Allow/deny list examples:

```bash
# Only allow shell commands containing the substring ALLOW_ME
./jorin --allow ALLOW_ME "Run the deployment script"

# Deny commands containing dangerous substrings
./jorin --deny "rm -rf" "passwd"
```

For more details about advanced REPL usage, shell policy and tool
permissions see the security and architecture docs:

- [security.md](security.md)
- [architecture.md](architecture.md)
- [reference.md](reference.md)

If you maintain repository-specific guidance, add an AGENTS.md file to the
project root — jorin will append the contents to the system prompt when run
from that repository.

You can also add personal Skills under ~/.jorin/skills. Each Skill should live
in its own folder with a SKILL.md that includes YAML frontmatter (name and
description). jorin will include the Skill descriptions in the system prompt.

Situations are executable context providers that emit prompt snippets. Create
them under ~/.jorin/situations or ./.jorin/situations (project-specific).
Each situation lives in its own folder with a SITUATION.yaml metadata file and
an executable referenced by the run field. When the executable prints output,
jorin wraps it in an XML-like tag matching the situation name.

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
