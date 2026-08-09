package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mobilelab-dev/mobilelab/pkg/plugin"
)

func main() {
	err := plugin.Serve(context.Background(), os.Stdin, os.Stdout, plugin.HandlerFunc(handle))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(_ context.Context, action string, input json.RawMessage, _ plugin.InvocationContext) (plugin.Result, error) {
	if action != "echo" {
		return plugin.Result{}, fmt.Errorf("unsupported action %q", action)
	}
	var received any
	if err := json.Unmarshal(input, &received); err != nil {
		return plugin.Result{}, fmt.Errorf("decode input: %w", err)
	}
	return plugin.Result{
		Message: "Echo plugin completed.",
		Output: map[string]any{
			"plugin":   "echo",
			"received": received,
		},
	}, nil
}
