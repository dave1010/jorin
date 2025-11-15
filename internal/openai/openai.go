package openai

import (
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
func ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	return DefaultLLM.ChatOnce(model, msgs, toolsList)
}

func ChatSession(model string, msgs []types.Message, pol *types.Policy) ([]types.Message, string, error) {
	toolsList := tools.ToolsManifest()
	reg := tools.Registry()
	for i := 0; i < 100; i++ {
		resp, err := ChatOnce(model, msgs, toolsList)
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
				// attempt to unmarshal args so we can display a concise preview
				var parsedArgs map[string]any
				if err := json.Unmarshal(tc.Function.Args, &parsedArgs); err != nil {
					// fallback: show trimmed raw args
					preview := tools.Preview(string(tc.Function.Args), 200)
					useColor := false
					if os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb" {
						useColor = true
					}
					if useColor {
						fmt.Fprintln(os.Stderr, "\x1b[36m"+tc.Function.Name+" "+preview+"\x1b[0m")
					} else {
						fmt.Fprintln(os.Stderr, tc.Function.Name+" "+preview)
					}
					// emit the same error behavior as before
					msgs = append(msgs, types.Message{
						Role:       "tool",
						Name:       tc.Function.Name,
						ToolCallID: tc.ID,
						Content:    `{"error":"bad arguments"}`,
					})
					continue
				}

				// build a concise, human-friendly preview based on tool type
				preview := ""
				switch tc.Function.Name {
				case "shell":
					if c, ok := parsedArgs["cmd"].(string); ok {
						preview = "$ " + tools.Preview(c, 200)
					} else {
						preview = "$ " + tools.Preview(string(tc.Function.Args), 200)
					}
				case "read_file":
					if p, ok := parsedArgs["path"].(string); ok {
						preview = "📄 " + p
					} else {
						preview = "📄 " + tools.Preview(string(tc.Function.Args), 200)
					}
				case "write_file":
					if p, ok := parsedArgs["path"].(string); ok {
						preview = "✏️ " + p
					} else {
						preview = "✏️ " + tools.Preview(string(tc.Function.Args), 200)
					}
				case "http_get":
					if u, ok := parsedArgs["url"].(string); ok {
						preview = "🌐 " + u
					} else {
						preview = "🌐 " + tools.Preview(string(tc.Function.Args), 200)
					}
				default:
					preview = tc.Function.Name + " " + tools.Preview(string(tc.Function.Args), 200)
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
					fmt.Fprintln(os.Stderr, col+preview+reset)
				} else {
					fmt.Fprintln(os.Stderr, preview)
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

				// call the tool with the parsed args
				out, _ := fn(parsedArgs, pol)
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
