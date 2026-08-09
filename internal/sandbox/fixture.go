package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadFixture(fixturesRoot, name string, variables map[string]string) ([]byte, error) {
	if name == "" {
		return nil, nil
	}
	root, err := filepath.Abs(fixturesRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve fixtures directory: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve fixtures directory symlinks: %w", err)
	}
	candidate, err := filepath.Abs(filepath.Join(root, filepath.Clean(name)))
	if err != nil {
		return nil, fmt.Errorf("resolve fixture %q: %w", name, err)
	}
	lexicalRelative, err := filepath.Rel(root, candidate)
	if err != nil || lexicalRelative == ".." || strings.HasPrefix(lexicalRelative, ".."+string(filepath.Separator)) || filepath.IsAbs(lexicalRelative) {
		return nil, fmt.Errorf("fixture %q escapes the fixtures directory", name)
	}
	resolvedCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return nil, fmt.Errorf("resolve fixture %q: %w", name, err)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedCandidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return nil, fmt.Errorf("fixture %q escapes the fixtures directory", name)
	}
	data, err := os.ReadFile(resolvedCandidate)
	if err != nil {
		return nil, fmt.Errorf("read fixture %q: %w", name, err)
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, fmt.Errorf("fixture %q is not valid JSON: %w", name, err)
	}
	data, err = json.Marshal(renderVariables(decoded, variables))
	if err != nil {
		return nil, fmt.Errorf("render fixture %q: %w", name, err)
	}
	return data, nil
}

func renderVariables(value any, variables map[string]string) any {
	switch typed := value.(type) {
	case string:
		for key, replacement := range variables {
			typed = strings.ReplaceAll(typed, "{{"+key+"}}", replacement)
		}
		return typed
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[key] = renderVariables(nested, variables)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, nested := range typed {
			result[index] = renderVariables(nested, variables)
		}
		return result
	default:
		return value
	}
}
