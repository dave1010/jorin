package plugins

import (
	"context"
	"fmt"
	"io"
)

func init() {
	p := &Plugin{
		Name:        "model-plugin",
		Description: "Provides /model command to show current model",
		Commands: map[string]CommandHandler{
			"plugins": pluginListHandler,
			"model":   modelHandler,
		},
	}
	RegisterPlugin(p)
}

func pluginListHandler(ctx context.Context, name string, args []string, raw string, out io.Writer, errOut io.Writer) (bool, error) {
	pls := ListPlugins()
	for _, p := range pls {
		if _, err := fmt.Fprintln(out, p.Name+": "+p.Description); err != nil {
			return true, err
		}
	}
	return true, nil
}

func modelHandler(ctx context.Context, name string, args []string, raw string, out io.Writer, errOut io.Writer) (bool, error) {
	m := Model()
	if m == "" {
		if _, err := fmt.Fprintln(out, "model: (unknown)"); err != nil {
			return true, err
		}
		return true, nil
	}
	if _, err := fmt.Fprintln(out, "model:", m); err != nil {
		return true, err
	}
	return true, nil
}
