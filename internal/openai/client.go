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
var DefaultLLM LLM = openAIClient{}

// UseResponses controls whether the OpenAI Responses API is used.
// It defaults to true, with Chat Completions available as a fallback.
var UseResponses = true

type openAIClient struct{}

func openAIBase() string {
	if b := os.Getenv("OPENAI_BASE_URL"); b != "" {
		return strings.TrimRight(b, "/")
	}
	return "https://api.openai.com"
}

func (o openAIClient) ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	if UseResponses {
		return o.chatOnceResponses(model, msgs, toolsList)
	}
	return o.chatOnceCompletions(model, msgs, toolsList)
}

func (o openAIClient) chatOnceCompletions(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
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

func (o openAIClient) chatOnceResponses(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	body := types.ResponseRequest{
		Model:      model,
		Input:      responseInputFromMessages(msgs),
		Tools:      responseToolsFromChatTools(toolsList),
		ToolChoice: "auto",
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
	var out types.Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return responseToChatResponse(&out), nil
}

func responseToolsFromChatTools(toolsList []types.Tool) []types.ResponseTool {
	if len(toolsList) == 0 {
		return nil
	}
	respTools := make([]types.ResponseTool, 0, len(toolsList))
	for _, tool := range toolsList {
		if tool.Type != "function" {
			continue
		}
		respTools = append(respTools, types.ResponseTool{
			Type:        "function",
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			Parameters:  tool.Function.Parameters,
		})
	}
	return respTools
}

func responseInputFromMessages(msgs []types.Message) []types.ResponseItem {
	items := []types.ResponseItem{}
	for _, msg := range msgs {
		if msg.Role == "tool" {
			items = append(items, types.ResponseItem{
				Type:   "function_call_output",
				CallID: msg.ToolCallID,
				Output: msg.Content,
			})
			continue
		}
		if msg.Content != "" {
			items = append(items, types.ResponseItem{
				Type: "message",
				Role: msg.Role,
				Content: []types.ResponseContent{
					{Type: "input_text", Text: msg.Content},
				},
			})
		}
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				items = append(items, types.ResponseItem{
					Type:      "function_call",
					Name:      tc.Function.Name,
					Arguments: tc.Function.Args,
					CallID:    tc.ID,
				})
			}
		}
	}
	return items
}

func responseToChatResponse(resp *types.Response) *types.ChatResponse {
	msg := types.Message{Role: "assistant"}
	for _, item := range resp.Output {
		switch item.Type {
		case "message":
			if item.Role != "assistant" {
				continue
			}
			for _, content := range item.Content {
				if content.Type == "output_text" {
					msg.Content += content.Text
				}
			}
		case "function_call":
			tc := types.ToolCall{
				ID:   item.CallID,
				Type: "function",
			}
			tc.Function.Name = item.Name
			tc.Function.Args = item.Arguments
			msg.ToolCalls = append(msg.ToolCalls, tc)
		}
	}
	return &types.ChatResponse{
		Choices: []types.Choice{
			{
				Message: msg,
			},
		},
	}
}
