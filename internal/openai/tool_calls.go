package openai

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dave1010/jorin/internal/tools"
	"github.com/dave1010/jorin/internal/types"
)

const colorReset = "\x1b[0m"

func handleToolCalls(calls []types.ToolCall, reg map[string]tools.ToolExec, pol *types.Policy) []types.Message {
	toolMsgs := make([]types.Message, 0, len(calls))
	for _, tc := range calls {
		parsedArgs, parsed := parseToolArgs(tc)
		preview := buildToolPreview(tc, parsedArgs, parsed)
		emitToolPreview(tc.Function.Name, preview)

		fn := reg[tc.Function.Name]
		if fn == nil {
			toolMsgs = append(toolMsgs, toolErrorMessage(tc, "unknown tool"))
			continue
		}
		if !parsed && parsedArgs == nil {
			parsedArgs = map[string]any{}
		}
		out, _ := fn(parsedArgs, pol)
		toolMsgs = append(toolMsgs, toolOutputMessage(tc, out))
	}
	return toolMsgs
}

func parseToolArgs(tc types.ToolCall) (map[string]any, bool) {
	var parsedArgs map[string]any
	if err := json.Unmarshal(tc.Function.Args, &parsedArgs); err == nil {
		return parsedArgs, true
	}
	var inner string
	if err := json.Unmarshal(tc.Function.Args, &inner); err != nil {
		return nil, false
	}
	if err := json.Unmarshal([]byte(inner), &parsedArgs); err == nil {
		return parsedArgs, true
	}
	return fallbackToolArgs(tc.Function.Name, inner)
}

func fallbackToolArgs(name string, inner string) (map[string]any, bool) {
	switch name {
	case "shell":
		return map[string]any{"cmd": inner}, true
	case "read_file":
		return map[string]any{"path": inner}, true
	case "write_file":
		return map[string]any{"path": inner, "text": ""}, true
	case "http_get":
		return map[string]any{"url": inner}, true
	default:
		return nil, false
	}
}

func buildToolPreview(tc types.ToolCall, parsedArgs map[string]any, parsed bool) string {
	if parsed {
		return previewFromParsedArgs(tc.Function.Name, parsedArgs, string(tc.Function.Args))
	}
	raw := toolPreviewRaw(tc.Function.Args)
	return previewFromRawArgs(tc.Function.Name, raw)
}

func previewFromParsedArgs(name string, args map[string]any, raw string) string {
	switch name {
	case "shell":
		return "$ " + tools.Preview(stringFromArg(args, "cmd", raw), 200)
	case "read_file":
		return "📄 " + stringFromArg(args, "path", tools.Preview(raw, 200))
	case "write_file":
		return "✏️ " + stringFromArg(args, "path", tools.Preview(raw, 200))
	case "http_get":
		return "🌐 " + stringFromArg(args, "url", tools.Preview(raw, 200))
	default:
		return name + " " + tools.Preview(raw, 200)
	}
}

func stringFromArg(args map[string]any, key string, fallback string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return fallback
}

func previewFromRawArgs(name string, raw string) string {
	switch name {
	case "shell":
		return "$ " + tools.Preview(raw, 200)
	case "read_file":
		return "📄 " + tools.Preview(raw, 200)
	case "write_file":
		return "✏️ " + tools.Preview(raw, 200)
	case "http_get":
		return "🌐 " + tools.Preview(raw, 200)
	default:
		return name + " " + tools.Preview(raw, 200)
	}
}

func toolPreviewRaw(raw json.RawMessage) string {
	var unquoted string
	if err := json.Unmarshal(raw, &unquoted); err == nil {
		return unquoted
	}
	return strings.TrimSpace(string(raw))
}

func emitToolPreview(toolName string, preview string) {
	if shouldColorizeOutput() {
		fmt.Fprintln(os.Stderr, toolColor(toolName)+preview+colorReset)
		return
	}
	fmt.Fprintln(os.Stderr, preview)
}

func shouldColorizeOutput() bool {
	return os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb"
}

func toolColor(name string) string {
	switch name {
	case "shell":
		return "\x1b[32m"
	case "read_file":
		return "\x1b[33m"
	case "write_file":
		return "\x1b[38;5;208m"
	default:
		return "\x1b[36m"
	}
}

func toolErrorMessage(tc types.ToolCall, message string) types.Message {
	return types.Message{
		Role:       "tool",
		Name:       tc.Function.Name,
		ToolCallID: tc.ID,
		Content:    fmt.Sprintf(`{"error":%q}`, message),
	}
}

func toolOutputMessage(tc types.ToolCall, out map[string]any) types.Message {
	b, _ := json.Marshal(out)
	return types.Message{
		Role:       "tool",
		Name:       tc.Function.Name,
		ToolCallID: tc.ID,
		Content:    string(b),
	}
}
