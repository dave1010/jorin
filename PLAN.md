# Jorin Refactoring and Improvement Plan

This document outlines a plan for refactoring and improving the Jorin codebase. The goal is to improve code quality, maintainability, and developer experience before adding new features.

## 1. Configuration and Dependency Injection [DONE]

- **Objective:** Decouple the `main` package from the `app` package and improve the way configuration is handled.

- **Specific Changes:**

    - **`cmd/jorin/main.go`:** [DONE]
        - In `main()`, instead of creating `app.Options`, create a dedicated configuration struct within the `main` package.
        - Create a new function `NewApp(config *Config)` in the `app` package that takes the configuration and returns a new `App` struct.
        - The `App` struct should contain the dependencies, such as the `agent.Agent` and the `repl.History`.
        - The `main` function will then call an `app.Run()` method on the `App` struct.

    - **`internal/app/app.go`:** [DONE]
        - Create a new `App` struct that holds the application's dependencies.
        - Create a `NewApp(opts Options)` function that initializes the `App` struct and its dependencies.
        - The `Run` function should be a method on the `App` struct and use the dependencies from the struct.
        - The `openai.DefaultAgent` should be instantiated in `NewApp` and passed to the `App` struct as an `agent.Agent` interface.

## 2. Componentization [DONE]

- **Objective:** Extract the "Ralph Wiggum" loop into its own component to improve separation of concerns.

- **Specific Changes:**

    - **`internal/app/app.go`:** [DONE]
        - Create a new package `internal/ralph` for the "Ralph Wiggum" loop.
        - Move the `runRalphLoop` function and the `ralphDone` function to the new `internal/ralph` package.
        - The `runPrompt` function in `internal/app/app.go` will then call the `ralph.Run` function.

## 3. Code Duplication [DONE]

- **Objective:** Centralize agent-related logic into the `internal/agent` package.

- **Specific Changes:**

    - **`cmd/jorin/agent.go`:** [DONE]
        - Move the `runAgent` function to a new file `internal/agent/run.go`.
        - Rename the function to `RunWithSystemPrompt`.
        - Remove the `cmd/jorin/agent.go` file.
    - **`internal/agent/agent.go`:** [DONE]
        - The existing `RunAgent` function will remain as it is.
    - **`cmd/jorin/agent_integration_test.go` and `cmd/jorin/main_agent_test.go`:** [DONE]
        - Move the tests to `internal/agent/run_test.go`.
        - Update the tests to use the new `RunWithSystemPrompt` function.
        - Remove `cmd/jorin/agent_integration_test.go` and `cmd/jorin/main_agent_test.go`.
        
## 4. Simplify main package [DONE]
- **Objective:** Reduce the amount of logic in the `main` package.

- **Specific Changes:**
    - **`cmd/jorin/main.go`:** [DONE]
        - Move the `resolvePromptMode`, `exitWithError` and `multi` flag helpers to a new `cmd/jorin/cli.go` file.
        - The `main` function should be the only function in `main.go` and should be responsible for parsing flags, creating the app, and running it.