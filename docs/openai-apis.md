# OpenAI APIs in Jorin

Jorin supports two different OpenAI API architectures: the standard **Chat Completions API** and the newer **Responses API**. This document explains their internal implementation details and key differences.

## Overview

Jorin abstracts these APIs behind the `LLM` interface in `internal/openai/client.go`:

```go
type LLM interface {
    ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error)
}
```

The choice of API is controlled by the `--use-responses-api` CLI flag, which selects between `completionsClient` and `responsesClient`.

---

## Chat Completions API (`/v1/chat/completions`)

This is the standard, well-known API used by most LLM applications.

### Request Structure
- **Endpoint**: `POST /v1/chat/completions`
- **Messages**: A flat list of messages with `role` and `content`.
- **Tools**: Defined as a nested structure where each tool has a `type` (always "function") and a `function` object containing the `name`, `description`, and `parameters`.

### State Management
The API is stateless. In every request, Jorin must send the entire conversation history (all previous messages and tool results) to provide context for the next turn.

---

## Responses API (`/v1/responses`)

This is a newer API designed for better session management and potentially more efficient iterative steps.

### Request Structure
- **Endpoint**: `POST /v1/responses`
- **Input**: A list of objects that can be of different types:
    - `message`: Standard chat message.
    - `function_call`: An item representing a tool being called (requires `call_id`).
    - `function_call_output`: The result of a tool execution (links to a `call_id`).
- **Instructions**: A top-level string typically used for the "system" prompt context.
- **Tools**: Unlike Chat Completions, the tool definition is **flattened**. `name`, `description`, and `parameters` are top-level fields within the tool object.
- **Previous Response ID**: A `previous_response_id` can be provided to reference the state of the conversation up to that point.

### State Management (Session Continuation)
The Responses API supports session continuation via `previous_response_id`.
- When Jorin receives a response, it stores the `ID` (e.g., `resp_...`).
- In the next request, Jorin sends `previous_response_id`.
- Jorin then only needs to send **new** messages or tool outputs that occurred *after* that response ID.

### Content Types
The Responses API is stricter about content types in the `input` slice:
- `user` messages must use `type: "input_text"`.
- `assistant` messages must use `type: "output_text"`.

---

## Key Differences Summary

| Feature | Chat Completions | Responses API |
| --- | --- | --- |
| **Endpoint** | `/v1/chat/completions` | `/v1/responses` |
| **Tool Definition** | Nested under `function` | Flattened (top-level fields) |
| **Session History** | Sent in full every request | Can use `previous_response_id` |
| **Tool Results** | `role: "tool"` message | `type: "function_call_output"` item |
| **Tool Calls** | `tool_calls` field in message | `type: "function_call"` item |
| **System Prompt** | `role: "system"` message | `instructions` top-level field |

## Debugging

To see exactly what is being sent and received by the Responses API, run Jorin with the `DEBUG=1` environment variable:

```bash
DEBUG=1 jorin --use-responses-api "your prompt"
```

This will output the raw JSON requests and responses to `stderr`.
