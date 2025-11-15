package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dave1010/jorin/internal/tools"
	"github.com/dave1010/jorin/internal/types"
)

// ChatOnce is a convenience wrapper that delegates to the package-level
// DefaultLLM implementation. Callers can swap DefaultLLM for a different
// provider in tests or to support other LLMs.
func ChatOnce(ctx context.Context, model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	return DefaultLLM.ChatOnce(ctx, model, msgs, toolsList)
}

func ChatSession(ctx context.Context, model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	toolsList := tools.ToolsManifest()
	reg := tools.Registry()
	for i := 0; i < 100; i++ {
		resp, err := ChatOnce(ctx, model, msgs, toolsList)
		if err != nil {
			return msgs, "", err
		}
		if len(resp.Choices) == 0 {
			return msgs, "", errors.New("no choices")
		}
		ch := resp.Choices[0]
		cm := ch.Message

		msgs = append(msgs, cm)

		if len(cm.ToolCalls) > 0 {
			for _, tc := range cm.ToolCalls {
				// show which tool is being called (trimmed)
				preview := tools.Preview(tc.Function.Args, 200)

				// determine prefix based on tool name
				prefix := tc.Function.Name
				switch tc.Function.Name {
				case "shell":
					prefix = "$"
				case "read_file":
					prefix = "@"
				case "write_file":
					prefix = "@w"
				}

				// decide whether to emit ANSI colors
				useColor := false
				if os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb" {
					useColor = true
				}

				if useColor {
					col := "\x1b[36m" // default cyan
					switch tc.Function.Name {
					case "shell":
						col = "\x1b[32m" // green
					case "read_file":
						col = "\x1b[33m" // yellow
					case "write_file":
						col = "\x1b[38;5;208m" // orange
					}
					reset := "\x1b[0m"
					fmt.Fprintln(os.Stderr, col+prefix+" "+preview+reset)
				} else {
					fmt.Fprintln(os.Stderr, prefix+" "+preview)
				}

				fn := reg[tc.Function.Name]
				if fn == nil {
					msgs = append(msgs, types.Message{
						Role:       "tool",
						Name:       tc.Function.Name,
						ToolCallID: tc.ID,
						Content:    `{"error":"unknown tool"}`,
					})
					continue
				}
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Args), &args); err != nil {
					msgs = append(msgs, types.Message{
						Role:       "tool",
						Name:       tc.Function.Name,
						ToolCallID: tc.ID,
						Content:    `{"error":"bad arguments"}`,
					})
					continue
				}
				out, _ := fn(args, pol)
				b, _ := json.Marshal(out)
				msgs = append(msgs, types.Message{
					Role:       "tool",
					Name:       tc.Function.Name,
					ToolCallID: tc.ID,
					Content:    string(b),
				})
			}
			continue
		}

		return msgs, cm.Content, nil
	}
	return msgs, "", errors.New("max turns reached")
}
