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

// CommandDef describes a command provided by a plugin. It may include a
// description, a handler, and nested subcommands.
type CommandDef struct {
	Description string
	Handler     CommandHandler
	Subcommands map[string]CommandDef
}

// Plugin describes a compiled-in plugin. It can register one or more command
// definitions keyed by command name.
type Plugin struct {
	Name        string
	Description string
	Commands    map[string]CommandDef
}

var (
	mu         sync.RWMutex
	plugins    []*Plugin
	commandMap = map[string]CommandHandler{}
	// metadata holds descriptions and subcommand info for help integration.
	metadata = map[string]struct {
		Desc string
		Sub  map[string]string
	}{}
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
	for name, def := range p.Commands {
		// ensure metadata entry
		m := metadata[name]
		if m.Sub == nil {
			m.Sub = map[string]string{}
		}
		if def.Description != "" {
			m.Desc = def.Description
		}
		metadata[name] = m
		if def.Handler != nil {
			commandMap[name] = def.Handler
		}
		// register subcommands
		for sn, sdef := range def.Subcommands {
			full := name + " " + sn
			if sdef.Handler != nil {
				commandMap[full] = sdef.Handler
			}
			if sdef.Description != "" {
				m := metadata[name]
				m.Sub[sn] = sdef.Description
				metadata[name] = m
			}
		}
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

// ListAllCommands returns a map of registered top-level command names to
// their descriptions.
func ListAllCommands() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	out := map[string]string{}
	for k, v := range metadata {
		out[k] = v.Desc
	}
	return out
}

// HelpForCommand returns the description and subcommands (name -> desc) for
// a top-level command, if present.
func HelpForCommand(name string) (desc string, subs map[string]string, ok bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := metadata[name]
	if !ok {
		return "", nil, false
	}
	// copy subs to avoid exposing internal map
	subsCopy := map[string]string{}
	for k, v := range m.Sub {
		subsCopy[k] = v
	}
	return m.Desc, subsCopy, true
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
