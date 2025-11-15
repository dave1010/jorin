package types

import "encoding/json"

// Messages and tool types

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name string          `json:"name"`
		Args json.RawMessage `json:"arguments"`
	} `json:"function"`
}

type Tool struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"` // JSON Schema
}

type ChatRequest struct {
	Model       string      `json:"model"`
	Messages    []Message   `json:"messages"`
	Tools       []Tool      `json:"tools,omitempty"`
	ToolChoice  interface{} `json:"tool_choice,omitempty"` // "auto"
	Temperature float32     `json:"temperature,omitempty"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Policy controls agent/tool behavior
type Policy struct {
	Readonly bool
	DryShell bool
	Allow    []string
	Deny     []string
	CWD      string
}

// Agent is the minimal interface used by the UI to interact with an LLM
// backend. Implementations (e.g., internal/agent) should satisfy this.
type Agent interface {
	ChatSession(model string, msgs []Message, pol *Policy) ([]Message, string, error)
}
