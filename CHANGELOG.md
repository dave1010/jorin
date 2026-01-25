# Changelog

This file summarizes notable project changes grouped into semantic version-style releases (most recent first).

## Unreleased

- API: fixed the OpenAI Responses API implementation to correctly map tools, handle function call IDs, and manage message history.
- API: added `DEBUG=1` environment variable to print raw JSON requests and responses for the Responses API to stderr.
- Refactor: moved `resolvePromptMode`, `exitWithError`, and `multi` flag helpers to `cmd/jorin/cli.go` to simplify `main.go`.
- Refactor: renamed `app.Options` to `app.Config` and updated `app.NewApp` to take a pointer to the config.
- Refactor: updated `app.App` and `ralph.Run` to use the injected `agent.Agent` dependency instead of static package-level calls.
- Refactor: simplify REPL wiring with option structs, helper functions, and typed command parse errors.
- Refactor: clarify OpenAI tool handling, CLI flag parsing, and tool registry helpers.
- Refactor: split app run flow and OpenAI tool-call helpers for clearer ownership.
- CLI: run Ralph Wiggum loop mode automatically until DONE or a configurable max-tries limit is reached, with per-iteration progress output.
- CLI: add --ralph to enable Ralph Wiggum loop instructions in the system prompt.
- CLI: support stdin prompts and executable Jorin scripts with shebang parsing plus argument forwarding.
- CLI: treat readable files passed as prompts as prompt files by default, with new --prompt/--prompt-file flags to control ambiguity.
- Installer: install binaries into /usr/local/bin when possible, otherwise ~/.local/bin, with PATH updates for user installs.
- Refactor: extracted system prompt providers into internal/prompt, moved REPL core and commands into internal/repl, and introduced internal/app for wiring.
- Tests: added an integration test suite that exercises system prompt composition and tool-call interactions via mock OpenAI servers.
- Tests: expanded tool-calling integration coverage for string argument fallback and unknown tool handling.
- Tests: added integration coverage for app runs, REPL command handling, and history output.

## v0.0.9 - 2026-01-11

- Docs: cleaned up README and usage docs, with updated Skills/Situations guidance and built-in situation list.
- Docs: added a comprehensive CLI reference and troubleshooting guide, and linked them from README and usage docs.
- Docs: added a Situations skill guide for creating and editing Situations.
- Prompt providers: added support for personal Skills under ~/.jorin/skills and ./.jorin/skills, adding their descriptions to the system prompt.
- Prompt providers: added executable Situations (from ~/.jorin/situations and ./.jorin/situations) that emit tagged prompt context.
- Prompt providers: moved runtime env/git/tooling context into built-in Situations under ./.jorin/situations.

## v0.0.8 — 2025-11-27

- CI: added Android build jobs and enforced formatting (gofmt) in CI workflows.
- CI: enhanced Android NDK support and broadened Android build matrix to cover additional targets.
- install.sh: detect Android environments and adjust installation steps accordingly.
- Release workflow: improved release steps to support Android NDK artifacts and cross-platform releases.

## v0.0.7 — 2025-11-26

- Plugin system: added a compiled-in plugin framework that lets the agent discover and expose plugin-provided commands. New commands and metadata are surfaced via the UI and command registry.
  - New CLI commands: /plugins (list installed plugins) and /model (inspect/set model providers).
  - Plugin subcommand metadata and help helpers make plugin commands appear in /help and improve discoverability.
  - The application can now select a model provider via the plugin registry (used by main), and a model plugin was adapted to the new helpers.
- Documentation and roadmap: documented the plugin system in usage docs and linked it from the README. The roadmap was revised and updated with plugin integration tasks (examples, integration tests, help formatting) and clearer phases/steps so contributors know the next priorities.
- UI: /help now includes plugin commands. Tests for command integration were added to ensure plugin commands behave correctly in the REPL/UI.
- Tests: unit tests added for the plugin registry, model provider wiring, and UI/commands to improve confidence in the new plugin code.
- Tooling/maintenance: small install/build convenience changes and tag-fetching improvements.

## v0.0.6 — 2025-11-15

- Prompt providers: The system prompt was refactored to be modular. Multiple
  PromptProvider implementations can be registered to contribute parts of the
  system prompt. Default providers include the immutable core instructions,
  AGENTS.md content, Git context (PWD and .git presence), OS environment
  (uname or GOOS/GOARCH), and tools on PATH.
- UI: The /debug command now prints the current composed system prompt so it's
  easy to inspect what is being sent to the model.
- Files: prompt providers moved to internal/ui/provider_*.go for improved
  separation and future plugin use.

- Internal: Introduced context-aware cancellation across the agent and OpenAI client interfaces. The Agent/LLM ChatSession/ChatOnce APIs now accept context.Context and package adapters/implementations were updated to propagate contexts. The HTTP client now uses http.NewRequestWithContext so in-flight requests can be canceled.
- REPL / UI: docs/usage.md updated to document that pressing ESC while the agent is running aborts the current request and immediately returns to the prompt (the background request continues but its output is ignored).
- API: openai adapter and DefaultAgent changed to accept contexts; agent interface updated accordingly.
- Build: bumped Go version to 1.24.0 and updated indirect dependencies (golang.org/x/sys v0.38.0, golang.org/x/term v0.37.0).

## v0.0.5 — 2025-11-15

- REPL: Added proper interactive line editing when running in a terminal. The REPL now uses a real line editor to provide left/right cursor movement, up/down history navigation, in-session history, and Ctrl-C abort support. When stdin/stdout are not terminals the REPL falls back to the previous scanner-based behavior for deterministic, testable input.
- UI: History from the provided History implementation is loaded into the interactive editor so past session commands are available via up-arrow.
- Internal: Introduced internal/ui/line_reader.go implementing a LineReader abstraction and wiring it into StartREPL.
- Tests: ensure existing tests still pass with the new line reader (scanner fallback preserved for non-ttys).

- Roadmap update: Phase 1 (package refactor and interfaces) marked COMPLETE.
- Phase 2 (REPL \u001f UI improvements) partially implemented. Completed items added to ROADMAP-PLAN.md and include:
  - internal/ui package with StartREPL supporting injected io.Reader/io.Writer and a Context-aware agent interface.
  - internal/ui/commands package with a deterministic slash-command parser (quoted args, escape prefix) and a Handler interface.
  - In-memory history implementation (internal/ui/history) and a default command handler supporting /help, /history and /debug.
  - REPL wiring: command parsing \u001f dispatch, escape-prefix handling, forwarding to agent, and basic tools registry support for leading '!' shell commands.
  - Unit tests covering command parsing and basic REPL flows.

## v0.0.4 — 2025-11-15

- Extracted an LLM client interface and introduced a dedicated OpenAI HTTP client implementation to decouple higher-level logic from request/response handling.
- Improved linting across the codebase: added golangci-lint configuration and Makefile targets, fixed numerous lint issues (checking Write/Setenv errors, ensuring resp.Body.Close is handled, silenced unused warnings in cmd), and tightened tests to validate error handling.
- REPL: allow running local shell commands when input starts with '!' (respects policy and dry-run modes).
- cmd: removed an unused chatOnce wrapper while keeping the chatSession wrapper; compacted tool output prefixes and added ANSI color handling for shell/read_file/write_file outputs.
- Refactor: removed shim types and moved more packages/types into internal packages; added/expanded internal tests to improve coverage.
- Documentation: added small docs for the OpenAI wrapper and other minor documentation updates.
- Miscellaneous housekeeping and small fixes.

## v0.0.3 — 2025-11-14

- Refactored package layout (moved internal types and OpenAI tools into internal packages).
- Added and expanded unit tests to improve coverage for core components.
- Cleaned up command imports and general code reorganization.
- Introduced CI workflow (GitHub Actions) and initial continuous-integration configuration.
- Miscellaneous project housekeeping and minor build improvements.

## v0.0.2 — 2025-11-13

- Documentation updates and README improvements.
- Added project license (MIT) and contributing guidance.
- Project layout adjustments and build/make improvements.
- Minor code and dependency alignment to prepare for more refactors.

## v0.0.1 — 2025-11-12

- Initial project scaffolding and layout (cmd/, modules, .gitignore).
- Added agent framework components and REPL support.
- Added support documents (AGENTS.md) and test scaffolding for tools.
- Several refactors and small feature additions to get the prototype working.
