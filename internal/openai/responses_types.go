package openai

import (
	"encoding/json"
)

type responsesRequest struct {
	Model              string      `json:"model"`
	Input              []any       `json:"input,omitempty"`
	Instructions       string      `json:"instructions,omitempty"`
	Tools              []any       `json:"tools,omitempty"`
	ToolChoice         interface{} `json:"tool_choice,omitempty"`
	Temperature        float32     `json:"temperature,omitempty"`
	PreviousResponseID string      `json:"previous_response_id,omitempty"`
}

type responsesResponse struct {
	ID     string                `json:"id"`
	Output []responsesOutputItem `json:"output"`
}

type responsesOutputItem struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	CallID    string          `json:"call_id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	Message   *outputMessage  `json:"message,omitempty"`
}

type outputMessage struct {
	Role    string          `json:"role"`
	Content []outputContent `json:"content"`
}

type outputContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type inputMessage struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type functionCallItem struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type functionCallOutputItem struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
}
