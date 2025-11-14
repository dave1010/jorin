package main

import (
	"github.com/dave1010/jorin/internal/tools"
	"github.com/dave1010/jorin/internal/types"
)

// Local ToolExec mirrors the original signature used in cmd package.
type ToolExec func(args map[string]any, cfg *types.Policy) (map[string]any, error)

func toolsManifest() (list []types.Tool) {
	return tools.ToolsManifest()
}

func registry() map[string]ToolExec {
	reg := tools.Registry()
	out := make(map[string]ToolExec)
	for k, v := range reg {
		// capture v
		fn := v
		out[k] = func(args map[string]any, p *types.Policy) (map[string]any, error) {
			return fn(args, p)
		}
	}
	return out
}
