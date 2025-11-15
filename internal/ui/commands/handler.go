package commands

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// History is a lightweight local interface used by handlers. It is purposely
// defined here to avoid importing the ui package and creating an import
// cycle. Any type implementing Add(line string) and List(limit int) []string
// can be passed.
type History interface {
	Add(line string)
	List(limit int) []string
}

// NewDefaultHandler returns a Handler that supports a few built-in commands
// and writes responses to out/errOut. It uses the provided History for
// '/history' command.
func NewDefaultHandler(out io.Writer, errOut io.Writer, hist History) Handler {
	return &defaultHandler{out: out, errOut: errOut, hist: hist}
}

type defaultHandler struct {
	out    io.Writer
	errOut io.Writer
	hist   History
}

var helpTopics = map[string]string{
	"repl": `repl - editing tips

The REPL provides a simple line editor. Here are some tips for editing multi-line
or longer text:

- To enter multiple lines, use your terminal's line continuation (e.g. write
  a sentence and press Enter). The REPL treats each submitted line as a new
  prompt to the agent. If you want to provide a single multi-line message to
  the agent, you can paste the full block into the prompt and submit once.

- Use the arrow keys (Up/Down) to navigate your history of previous inputs.
  If the program was started with a history configured, past sessions will be
  available as well.

- To include literal leading slashes (e.g. to start your message with
  "/help"), prefix with an escape character defined in the config (by
  default the REPL attempts to detect escapes). If you find your leading
  slash is interpreted as a command, try escaping it.

- For shell commands, you can invoke them with a leading '!' (for example
  '!ls -la'). This uses the configured shell tool. Be cautious with destructive
  commands.

- While the agent is running, press the ESC key to abort and return to the
  prompt immediately. This is a best-effort UI feature: the background agent
  request will continue but its output will be ignored.

- Use /history to review recent inputs, and /help repl to show this topic again.
`,
}

func (d *defaultHandler) Handle(ctx context.Context, cmd Command) (bool, error) {
	switch cmd.Name {
	case "debug":
		// print nothing here; caller can call ui.SystemPrompt if needed
		if _, err := fmt.Fprintln(d.errOut, "debug: use /debug to print runtime/system info"); err != nil {
			return false, err
		}
		return true, nil
	case "help":
		// If a specific topic was requested, show it. Otherwise list commands
		// and available help topics.
		if len(cmd.Args) > 0 {
			topic := strings.ToLower(cmd.Args[0])
			if content, ok := helpTopics[topic]; ok {
				if _, err := fmt.Fprintln(d.out, content); err != nil {
					return false, err
				}
				return true, nil
			}
			if _, err := fmt.Fprintln(d.errOut, "unknown help topic:", topic); err != nil {
				return false, err
			}
			// fallthrough to list available topics
		}
		if _, err := fmt.Fprintln(d.out, "Available commands: /help [topic], /history [n], /debug"); err != nil {
			return false, err
		}
		// list help topics
		topics := []string{}
		for k := range helpTopics {
			topics = append(topics, k)
		}
		if len(topics) > 0 {
			if _, err := fmt.Fprintln(d.out, "Help topics:"); err != nil {
				return false, err
			}
			for _, t := range topics {
				if _, err := fmt.Fprintln(d.out, "  "+t); err != nil {
					return false, err
				}
			}
		}
		return true, nil
	case "history":
		limit := 0
		if len(cmd.Args) > 0 {
			if v, err := strconv.Atoi(cmd.Args[0]); err == nil {
				limit = v
			}
		}
		if d.hist == nil {
			if _, err := fmt.Fprintln(d.errOut, "history not available"); err != nil {
				return false, err
			}
			return true, nil
		}
		list := d.hist.List(limit)
		for i := range list {
			if _, err := fmt.Fprintln(d.out, list[i]); err != nil {
				return false, err
			}
		}
		return true, nil
	default:
		// unknown commands result in a friendly message
		if _, err := fmt.Fprintln(d.errOut, "unknown command:", cmd.Raw); err != nil {
			return false, err
		}
		return true, nil
	}
}
