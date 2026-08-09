package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	pluginruntime "github.com/mobilelab-dev/mobilelab/internal/plugins"
	protocol "github.com/mobilelab-dev/mobilelab/pkg/plugin"
)

func (r Runner) pluginCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mobilelab plugin list | inspect <name> [--json] | run <name> <action> [--input file.json] [--timeout 30s] [--output file.json]")
	}
	switch args[0] {
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("usage: mobilelab plugin list")
		}
		return r.listPlugins()
	case "inspect":
		if len(args) < 2 || len(args) > 3 || (len(args) == 3 && args[2] != "--json") {
			return fmt.Errorf("usage: mobilelab plugin inspect <name> [--json]")
		}
		return r.inspectPlugin(args[1], len(args) == 3)
	case "run":
		return r.runPlugin(ctx, args[1:])
	default:
		return fmt.Errorf("unknown plugin command %q; use list, inspect, or run", args[0])
	}
}

func (r Runner) listPlugins() error {
	catalog := pluginruntime.Catalog{ProjectDir: r.Dir}
	descriptors, issues, err := catalog.Discover()
	if err != nil {
		return err
	}
	if len(descriptors) == 0 && len(issues) == 0 {
		fmt.Fprintf(r.Out, "No plugins found in %s.\n", catalog.Root())
		return nil
	}
	if len(descriptors) > 0 {
		fmt.Fprintln(r.Out, "Project plugins:")
		for _, descriptor := range descriptors {
			actions := make([]string, 0, len(descriptor.Manifest.Actions))
			for _, action := range descriptor.Manifest.Actions {
				actions = append(actions, action.Name)
			}
			fmt.Fprintf(r.Out, "- %-24s v%-12s %s\n  actions: %s\n", descriptor.Manifest.Name, descriptor.Manifest.Version, descriptor.Manifest.Description, strings.Join(actions, ", "))
		}
	}
	for _, issue := range issues {
		fmt.Fprintf(r.Err, "! Invalid plugin at %s: %v\n", issue.Path, issue.Err)
	}
	if len(issues) > 0 {
		return fmt.Errorf("%d invalid plugin(s) found", len(issues))
	}
	return nil
}

func (r Runner) inspectPlugin(name string, jsonOutput bool) error {
	descriptor, err := (pluginruntime.Catalog{ProjectDir: r.Dir}).Load(name)
	if err != nil {
		return fmt.Errorf("inspect plugin %q: %w", name, err)
	}
	if jsonOutput {
		encoder := json.NewEncoder(r.Out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(descriptor)
	}
	fmt.Fprintf(r.Out, "Plugin: %s\nVersion: %s\nProtocol: %s\nDescription: %s\nExecutable: %s\nSHA-256: %s\nActions:\n", descriptor.Manifest.Name, descriptor.Manifest.Version, descriptor.Manifest.APIVersion, descriptor.Manifest.Description, descriptor.Executable, descriptor.SHA256)
	for _, action := range descriptor.Manifest.Actions {
		fmt.Fprintf(r.Out, "- %-20s %s\n", action.Name, action.Description)
	}
	return nil
}

func (r Runner) runPlugin(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mobilelab plugin run <name> <action> [--input file.json] [--timeout 30s] [--output file.json]")
	}
	name, action := args[0], args[1]
	var inputPath, outputPath string
	timeout := pluginruntime.DefaultTimeout
	for index := 2; index < len(args); index++ {
		flag := args[index]
		if index+1 >= len(args) {
			return fmt.Errorf("option %q requires a value", flag)
		}
		value := args[index+1]
		index++
		switch flag {
		case "--input":
			inputPath = value
		case "--output":
			outputPath = value
		case "--timeout":
			parsed, err := time.ParseDuration(value)
			if err != nil || parsed <= 0 || parsed > 5*time.Minute {
				return fmt.Errorf("plugin timeout must be greater than zero and no more than 5m")
			}
			timeout = parsed
		default:
			return fmt.Errorf("unknown plugin run option %q", flag)
		}
	}
	input := json.RawMessage(`{}`)
	if inputPath != "" {
		data, err := readLimitedFile(inputPath, protocol.MaxMessageBytes)
		if err != nil {
			return fmt.Errorf("read plugin input: %w", err)
		}
		if !json.Valid(data) {
			return fmt.Errorf("plugin input %q must contain valid JSON", inputPath)
		}
		input = data
	}
	descriptor, err := (pluginruntime.Catalog{ProjectDir: r.Dir}).Load(name)
	if err != nil {
		return fmt.Errorf("load plugin %q: %w", name, err)
	}
	response, err := (pluginruntime.Executor{Process: r.PluginProcess}).Execute(ctx, descriptor, action, input, timeout, Version)
	if err != nil {
		return err
	}
	if response.Message != "" {
		fmt.Fprintln(r.Err, response.Message)
	}
	formatted := []byte("null\n")
	if len(response.Output) > 0 {
		var buffer bytes.Buffer
		if err := json.Indent(&buffer, response.Output, "", "  "); err != nil {
			return fmt.Errorf("format plugin output: %w", err)
		}
		buffer.WriteByte('\n')
		formatted = buffer.Bytes()
	}
	if outputPath == "" {
		_, err = r.Out.Write(formatted)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create plugin output directory: %w", err)
	}
	if err := writePrivateFile(outputPath, formatted); err != nil {
		return fmt.Errorf("write plugin output %q: %w", outputPath, err)
	}
	return nil
}

func writePrivateFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func readLimitedFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds %d bytes", limit)
	}
	return data, nil
}
