package plugins

import (
	"context"
	"io"
	"sync"
)

// CommandHandler is the signature for handling a slash command registered by a
// plugin. It receives the command name, args and raw input and writers for
// stdout/stderr. Return handled=true if the default REPL should not forward
// the original line to the model.
type CommandHandler func(ctx context.Context, name string, args []string, raw string, out io.Writer, errOut io.Writer) (bool, error)

// Plugin describes a compiled-in plugin. It can register one or more command
// handlers keyed by command name.
type Plugin struct {
	Name        string
	Description string
	Commands    map[string]CommandHandler
}

var (
	mu         sync.RWMutex
	plugins    []*Plugin
	commandMap = map[string]CommandHandler{}
	// modelProvider can be set by the host (eg. the UI) so plugins can access
	// the current model name.
	modelProvider func() string
)

// RegisterPlugin registers a plugin and its commands. If a command name
// conflicts with an existing registered command, the latest registration
// wins.
func RegisterPlugin(p *Plugin) {
	mu.Lock()
	defer mu.Unlock()
	plugins = append(plugins, p)
	for k, h := range p.Commands {
		commandMap[k] = h
	}
}

// ListPlugins returns the list of registered plugins in registration order.
func ListPlugins() []*Plugin {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Plugin, len(plugins))
	copy(out, plugins)
	return out
}

// LookupCommand returns a command handler for the given name, if any.
func LookupCommand(name string) (CommandHandler, bool) {
	mu.RLock()
	defer mu.RUnlock()
	h, ok := commandMap[name]
	return h, ok
}

// SetModelProvider sets a callback used by plugins to obtain the current
// model name. Host (UI) should set this so plugins can read model info.
func SetModelProvider(f func() string) {
	mu.Lock()
	defer mu.Unlock()
	modelProvider = f
}

// Model returns the current model name from the provider or empty string if
// not available.
func Model() string {
	mu.RLock()
	defer mu.RUnlock()
	if modelProvider == nil {
		return ""
	}
	return modelProvider()
}
