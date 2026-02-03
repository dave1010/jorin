---
name: situations
description: "Use when creating or editing Situations - Explains required files, metadata, and run script behavior."
---

## Purpose

Situations capture small, runnable context snippets that are appended to prompts under a `## Situations` section.

## Location and layout

- Place each Situation in its own directory under `.jorin/situations/` (repo-scoped) or `~/.jorin/situations/` (user-scoped).
- Each directory must include a `SITUATION.yaml` file and a runnable script referenced by `run`.

Example layout:

```
.jorin/situations/<situation-name>/
  SITUATION.yaml
  run
```

## SITUATION.yaml fields

- `name`: Identifier used for output tags like `<name>...</name>`. Defaults to the directory name if omitted.
- `description`: Brief human description of the situation’s purpose.
- `run`: Relative path (typically `run`) to the executable script.

If `run` is missing, the situation is ignored.

## Run script expectations

- Prefer `#!/usr/bin/env bash` with `set -euo pipefail` for reliability.
- Emit concise, plain-text output to stdout.
- Keep output stable and low-noise because it is inserted directly into prompt context.
- The runner executes from the working directory and sets `JORIN_PWD` to that path; use it if you need an absolute reference.
- Ensure the script is executable; if direct execution fails, it is retried with `bash` and `sh`.

## Editing guidelines

- Update the `description` when the output meaning changes.
- Keep the script fast and deterministic; avoid network calls or long-running operations.
- Validate locally by running the script directly and checking the output formatting.
