package commands

import (
	"context"
	"fmt"
	"io"
	"strconv"
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

func (d *defaultHandler) Handle(ctx context.Context, cmd Command) (bool, error) {
	switch cmd.Name {
	case "debug":
		// print nothing here; caller can call ui.SystemPrompt if needed
		fmt.Fprintln(d.errOut, "debug: use /debug to print runtime/system info")
		return true, nil
	case "help":
		fmt.Fprintln(d.out, "Available commands: /help, /history [n], /debug")
		return true, nil
	case "history":
		limit := 0
		if len(cmd.Args) > 0 {
			if v, err := strconv.Atoi(cmd.Args[0]); err == nil {
				limit = v
			}
		}
		if d.hist == nil {
			fmt.Fprintln(d.errOut, "history not available")
			return true, nil
		}
		list := d.hist.List(limit)
		for i := range list {
			fmt.Fprintln(d.out, list[i])
		}
		return true, nil
	default:
		// unknown commands result in a friendly message
		fmt.Fprintln(d.errOut, "unknown command:", cmd.Raw)
		return true, nil
	}
}
