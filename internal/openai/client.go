package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dave1010/jorin/internal/types"
)

// LLM is an interface representing a language model provider that can
// produce a single chat completion. This abstraction lets us add other
// LLM backends later without duplicating session orchestration logic.
type LLM interface {
	ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error)
}

// DefaultLLM is the package-level LLM implementation used by the
// convenience functions in this package. It defaults to the OpenAI
// HTTP client implementation below.
var DefaultLLM LLM = completionsClient{}

type completionsClient struct{}

type responsesClient struct{}

func openAIBase() string {
	if b := os.Getenv("OPENAI_BASE_URL"); b != "" {
		return strings.TrimRight(b, "/")
	}
	return "https://api.openai.com"
}

func (o completionsClient) ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	body := types.ChatRequest{
		Model:      model,
		Messages:   msgs,
		Tools:      toolsList,
		ToolChoice: "auto",
	}
	j, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", openAIBase()+"/v1/chat/completions", bytes.NewReader(j))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(b))
	}
	var out types.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (o responsesClient) ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	input := []any{}
	var instructions string
	var previousResponseID string

	// Find the last message with a ResponseID to use as previousResponseID
	lastWithID := -1
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].ResponseID != "" {
			previousResponseID = msgs[i].ResponseID
			lastWithID = i
			break
		}
	}

	for i, m := range msgs {
		if m.Role == "system" {
			instructions += m.Content + "\n"
			continue
		}

		// If we found a previousResponseID, skip messages that are already part of it
		if i <= lastWithID {
			continue
		}

		if m.Role == "tool" {
			input = append(input, functionCallOutputItem{
				Type:   "function_call_output",
				CallID: m.ToolCallID,
				Output: m.Content,
			})
			continue
		}

		input = append(input, inputMessage{
			Type:    "message",
			Role:    m.Role,
			Content: m.Content,
		})

		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				input = append(input, functionCallItem{
					Type:      "function_call",
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Args,
				})
			}
		}
	}

	// Map types.Tool to responses API tool format
	var tools []any
	for _, t := range toolsList {
		tools = append(tools, t)
	}

	body := responsesRequest{
		Model:              model,
		Input:              input,
		Instructions:       instructions,
		Tools:              tools,
		ToolChoice:         "auto",
		PreviousResponseID: previousResponseID,
	}
	j, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", openAIBase()+"/v1/responses", bytes.NewReader(j))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(b))
	}

	var r responsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return mapResponseToChatResponse(&r), nil
}

func mapResponseToChatResponse(r *responsesResponse) *types.ChatResponse {
	res := &types.ChatResponse{
		ID: r.ID,
	}

	msg := types.Message{
		Role: "assistant",
	}

	for _, o := range r.Output {
		switch o.Type {
		case "message":
			if o.Message != nil {
				for _, c := range o.Message.Content {
					if c.Type == "output_text" || c.Type == "text" {
						msg.Content += c.Text
					}
				}
			}
		case "function_call":
			msg.ToolCalls = append(msg.ToolCalls, types.ToolCall{
				ID:   o.ID,
				Type: "function",
				Function: struct {
					Name string          `json:"name"`
					Args json.RawMessage `json:"arguments"`
				}{
					Name: o.Name,
					Args: o.Arguments,
				},
			})
		}
	}

	res.Choices = []types.Choice{
		{
			Message:      msg,
			FinishReason: "stop", // Assume stop if we got everything
		},
	}

	return res
}
