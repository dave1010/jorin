# Changelog

This file summarizes notable project changes grouped into semantic version-style releases (most recent first).

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
