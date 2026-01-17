package commands

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

// Command represents a parsed slash command line.
type Command struct {
	Raw  string
	Name string
	Args []string
}

// Handler handles parsed commands. Return handled=true to indicate the
// REPL should not forward the original line to the agent.
type Handler interface {
	Handle(ctx context.Context, cmd Command) (handled bool, err error)
}

// Parse takes a raw line and returns a Command if it is a slash command
// (starting with prefix). It supports single- and double-quoted args and
// a simple escape mechanism: if the line starts with EscapePrefix+Prefix
// the escape is removed and the line is not treated as a command.
func Parse(line string, prefix string, escapePrefix string) (Command, error) {
	c := Command{Raw: line}
	if prefix == "" {
		prefix = "/"
	}
	if escapePrefix == "" {
		escapePrefix = "\\"
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return c, errors.New("empty")
	}
	// escaped prefix, e.g. "\/help" -> "/help" (not a command)
	if strings.HasPrefix(line, escapePrefix+prefix) {
		c.Raw = strings.TrimPrefix(line, escapePrefix)
		return c, errors.New("escaped")
	}
	if !strings.HasPrefix(line, prefix) {
		return c, errors.New("not a command")
	}
	// remove leading prefix
	s := strings.TrimPrefix(line, prefix)
	// split into name and args using a regexp that respects quotes
	re := regexp.MustCompile(`'[^']*'|"[^"]*"|\S+`)
	parts := re.FindAllString(s, -1)
	if len(parts) == 0 {
		return c, errors.New("empty command")
	}
	// extract name (first token) and strip quotes from args
	c.Name = parts[0]
	if len(parts) > 1 {
		for _, p := range parts[1:] {
			p = strings.TrimSpace(p)
			if len(p) >= 2 && ((p[0] == '\'' && p[len(p)-1] == '\'') || (p[0] == '"' && p[len(p)-1] == '"')) {
				p = p[1 : len(p)-1]
			}
			c.Args = append(c.Args, p)
		}
	}
	return c, nil
}
