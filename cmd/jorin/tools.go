package main

import (
	"github.com/dave1010/jorin/internal/tools"
	"github.com/dave1010/jorin/internal/types"
)

// Local ToolExec mirrors the original signature used in cmd package.
type ToolExec func(args map[string]any, cfg *Policy) (map[string]any, error)

func toolsManifest() (list []Tool) {
	in := tools.ToolsManifest()
	out := make([]Tool, 0, len(in))
	for _, t := range in {
		out = append(out, Tool{Type: t.Type, Function: ToolFunction{Name: t.Function.Name, Description: t.Function.Description, Parameters: t.Function.Parameters}})
	}
	return out
}

func registry() map[string]ToolExec {
	reg := tools.Registry()
	out := make(map[string]ToolExec)
	for k, v := range reg {
		// capture v to avoid loop variable issue
		fn := v
		out[k] = func(args map[string]any, p *Policy) (map[string]any, error) {
			return fn(args, (*types.Policy)(p))
		}
	}
	return out
}
