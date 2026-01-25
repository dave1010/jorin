package commands

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/dave1010/jorin/internal/plugins"
)

// History is a lightweight local interface used by handlers. It is purposely
// defined here to avoid importing the repl package and creating an import
// cycle. Any type implementing Add(line string) and List(limit int) []string
// can be passed.
type History interface {
	Add(line string)
	List(limit int) []string
}

// NewDefaultHandler returns a Handler that supports a few built-in commands
// and writes responses to out/errOut. It uses the provided History for
// '/history' command. getSystemPrompt is a callback that returns the current
// system prompt; it's supplied by the caller to avoid an import cycle with the
// prompt package.
func NewDefaultHandler(out io.Writer, errOut io.Writer, hist History, getSystemPrompt func() string) Handler {
	return &defaultHandler{out: out, errOut: errOut, hist: hist, sysPrompt: getSystemPrompt}
}

type defaultHandler struct {
	out       io.Writer
	errOut    io.Writer
	hist      History
	sysPrompt func() string
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

- Use /history to review recent inputs, and /help repl to show this topic again.
`,
}

func (d *defaultHandler) Handle(ctx context.Context, cmd Command) (bool, error) {
	if handled, err := d.handlePluginCommand(ctx, cmd); handled || err != nil {
		return handled, err
	}
	switch cmd.Name {
	case "debug":
		return d.handleDebug()
	case "help":
		return d.handleHelp(cmd)
	case "history":
		return d.handleHistory(cmd)
	default:
		return d.handleUnknown(cmd)
	}
}

func (d *defaultHandler) handlePluginCommand(ctx context.Context, cmd Command) (bool, error) {
	if h, ok := plugins.LookupCommand(cmd.Name); ok {
		return h(ctx, cmd.Name, cmd.Args, cmd.Raw, d.out, d.errOut)
	}
	return false, nil
}

func (d *defaultHandler) handleDebug() (bool, error) {
	if d.sysPrompt == nil {
		if _, err := fmt.Fprintln(d.errOut, "debug: system prompt not available"); err != nil {
			return false, err
		}
		return true, nil
	}
	if _, err := fmt.Fprintln(d.errOut, d.sysPrompt()); err != nil {
		return false, err
	}
	return true, nil
}

func (d *defaultHandler) handleHelp(cmd Command) (bool, error) {
	if len(cmd.Args) > 0 {
		topic := strings.ToLower(cmd.Args[0])
		handled, err := d.writeHelpTopic(topic)
		if err != nil {
			return false, err
		}
		if handled {
			return true, nil
		}
		if _, err := fmt.Fprintln(d.errOut, "unknown help topic:", topic); err != nil {
			return false, err
		}
	}
	return d.writeHelpIndex()
}

func (d *defaultHandler) writeHelpTopic(topic string) (bool, error) {
	if content, ok := helpTopics[topic]; ok {
		if _, err := fmt.Fprintln(d.out, content); err != nil {
			return false, err
		}
		return true, nil
	}
	desc, subs, ok := plugins.HelpForCommand(topic)
	if !ok {
		return false, nil
	}
	if desc != "" {
		if _, err := fmt.Fprintln(d.out, desc); err != nil {
			return false, err
		}
	}
	if len(subs) > 0 {
		if _, err := fmt.Fprintln(d.out, "Subcommands:"); err != nil {
			return false, err
		}
		for sn, sdesc := range subs {
			if _, err := fmt.Fprintln(d.out, "  "+sn+": "+sdesc); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (d *defaultHandler) writeHelpIndex() (bool, error) {
	if _, err := fmt.Fprintln(d.out, "Available commands: /help [topic], /history [n], /debug"); err != nil {
		return false, err
	}
	topics := sortedHelpTopics()
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
	if err := d.writePluginCommands(); err != nil {
		return false, err
	}
	return true, nil
}

func (d *defaultHandler) writePluginCommands() error {
	cmds := plugins.ListAllCommands()
	if len(cmds) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(d.out, "Plugin commands:"); err != nil {
		return err
	}
	names := make([]string, 0, len(cmds))
	for name := range cmds {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		desc := cmds[name]
		if desc == "" {
			if _, err := fmt.Fprintln(d.out, "  /"+name); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintln(d.out, "  /"+name+" - "+desc); err != nil {
			return err
		}
	}
	return nil
}

func (d *defaultHandler) handleHistory(cmd Command) (bool, error) {
	if d.hist == nil {
		if _, err := fmt.Fprintln(d.errOut, "history not available"); err != nil {
			return false, err
		}
		return true, nil
	}
	limit := parseHistoryLimit(cmd.Args)
	list := d.hist.List(limit)
	for _, line := range list {
		if _, err := fmt.Fprintln(d.out, line); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (d *defaultHandler) handleUnknown(cmd Command) (bool, error) {
	if _, err := fmt.Fprintln(d.errOut, "unknown command:", cmd.Raw); err != nil {
		return false, err
	}
	return true, nil
}

func parseHistoryLimit(args []string) int {
	if len(args) == 0 {
		return 0
	}
	limit, err := strconv.Atoi(args[0])
	if err != nil {
		return 0
	}
	return limit
}

func sortedHelpTopics() []string {
	topics := make([]string, 0, len(helpTopics))
	for k := range helpTopics {
		topics = append(topics, k)
	}
	sort.Strings(topics)
	return topics
}
