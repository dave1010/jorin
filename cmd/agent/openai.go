package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func openAIBase() string {
	if b := os.Getenv("OPENAI_BASE_URL"); b != "" {
		return strings.TrimRight(b, "/")
	}
	return "https://api.openai.com"
}

func chatOnce(model string, msgs []Message, tools []Tool) (*ChatResponse, error) {
	body := ChatRequest{
		Model:      model,
		Messages:   msgs,
		Tools:      tools,
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
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(b))
	}
	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func chatSession(model string, msgs []Message, pol *Policy) ([]Message, string, error) {
	tools := toolsManifest()
	reg := registry()
	for i := 0; i < 100; i++ {
		resp, err := chatOnce(model, msgs, tools)
		if err != nil {
			return msgs, "", err
		}
		if len(resp.Choices) == 0 {
			return msgs, "", errors.New("no choices")
		}
		ch := resp.Choices[0]
		cm := ch.Message

		// append assistant message
		msgs = append(msgs, cm)

		// If tool calls present, execute them serially and loop
		if len(cm.ToolCalls) > 0 {
			for _, tc := range cm.ToolCalls {
				// show which tool is being called (trimmed)
				fmt.Fprintln(os.Stderr, "CALL TOOL:", tc.Function.Name, preview(tc.Function.Args, 200))
				fn := reg[tc.Function.Name]
				if fn == nil {
					msgs = append(msgs, Message{
						Role:       "tool",
						Name:       tc.Function.Name,
						ToolCallID: tc.ID,
						Content:    `{"error":"unknown tool"}`,
					})
					continue
				}
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Args), &args); err != nil {
					msgs = append(msgs, Message{
						Role:       "tool",
						Name:       tc.Function.Name,
						ToolCallID: tc.ID,
						Content:    `{"error":"bad arguments"}`,
					})
					continue
				}
				out, _ := fn(args, pol)
				b, _ := json.Marshal(out)
				msgs = append(msgs, Message{
					Role:       "tool",
					Name:       tc.Function.Name,
					ToolCallID: tc.ID,
					Content:    string(b),
				})
			}
			// continue the loop so model sees tool outputs and can respond
			continue
		}

		// otherwise final assistant content (no tool calls)
		return msgs, cm.Content, nil
	}
	return msgs, "", errors.New("max turns reached")
}
