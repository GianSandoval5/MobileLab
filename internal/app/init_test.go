package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/config"
)

func TestInitializeCreatesUsableEnvironment(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pubspec.yaml"), []byte("name: demo"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Initialize(root)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if len(result.Detected) != 1 || result.Detected[0].Name != "Flutter" {
		t.Fatalf("unexpected detection: %v", result.Detected)
	}
	generatedConfig, err := config.Load(result.ConfigPath)
	if err != nil {
		t.Fatalf("generated config cannot be loaded: %v", err)
	}
	if generatedConfig.SchemaVersion != 1 {
		t.Fatalf("generated config schema version = %d", generatedConfig.SchemaVersion)
	}
	for _, path := range []string{
		"mobilelab/.gitignore",
		"mobilelab/fixtures/profile.json",
		"mobilelab/scenarios/profile.yaml",
		"mobilelab/mocks",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Errorf("missing generated path %s: %v", path, err)
		}
	}
	scenarioData, err := os.ReadFile(filepath.Join(root, "mobilelab", "scenarios", "profile.yaml"))
	if err != nil || !strings.HasPrefix(string(scenarioData), "schema_version: 1\n") {
		t.Fatalf("generated scenario is not schema v1: %q %v", scenarioData, err)
	}
}

func TestInitializeDoesNotOverwriteConfig(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, config.DefaultFilename)
	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Initialize(root)
	if err == nil || !strings.Contains(err.Error(), "did not overwrite") {
		t.Fatalf("expected no-overwrite error, got %v", err)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil || string(data) != "original" {
		t.Fatalf("existing config changed: %q, %v", data, readErr)
	}
}
