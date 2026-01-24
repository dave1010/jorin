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

// Responses API types

type ResponseTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type ResponseRequest struct {
	Model       string         `json:"model"`
	Input       any            `json:"input"`
	Tools       []ResponseTool `json:"tools,omitempty"`
	ToolChoice  interface{}    `json:"tool_choice,omitempty"`
	Temperature float32        `json:"temperature,omitempty"`
}

type Response struct {
	Output []ResponseItem `json:"output"`
}

type ResponseItem struct {
	ID        string            `json:"id,omitempty"`
	Type      string            `json:"type"`
	Role      string            `json:"role,omitempty"`
	Status    string            `json:"status,omitempty"`
	Content   []ResponseContent `json:"content,omitempty"`
	Name      string            `json:"name,omitempty"`
	Arguments json.RawMessage   `json:"arguments,omitempty"`
	CallID    string            `json:"call_id,omitempty"`
	Output    string            `json:"output,omitempty"`
}

type ResponseContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type ResponseFunctionCallItem struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
	CallID    string          `json:"call_id"`
}

type ResponseFunctionCallOutputItem struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
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
