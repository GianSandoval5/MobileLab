package migration

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/config"
)

func TestPlanPreflightsThenAtomicallyMigratesLegacyProject(t *testing.T) {
	root := legacyProject(t)
	plan, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.ChangeCount() != 2 {
		t.Fatalf("change count = %d, want 2: %#v", plan.ChangeCount(), plan.Documents)
	}
	configPath := filepath.Join(root, config.DefaultFilename)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(before), "schema_version") {
		t.Fatal("planning unexpectedly modified the project")
	}
	if err := plan.Apply(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{configPath, filepath.Join(root, "mobilelab", "scenarios", "smoke.yaml")} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), "schema_version: 1\n") {
			t.Fatalf("%s was not migrated: %s", path, data)
		}
	}
	scenarioData, err := os.ReadFile(filepath.Join(root, "mobilelab", "scenarios", "smoke.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(scenarioData), "# preserved comment") {
		t.Fatalf("scenario comment was not preserved: %s", scenarioData)
	}
	configInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && configInfo.Mode().Perm() != 0o640 {
		t.Fatalf("configuration mode = %o, want 640", configInfo.Mode().Perm())
	}
	current, err := Build(root)
	if err != nil || current.ChangeCount() != 0 {
		t.Fatalf("migrated project is not current: %#v %v", current, err)
	}
}

func TestPlanRejectsInvalidScenarioBeforeAnyWrite(t *testing.T) {
	root := legacyProject(t)
	broken := filepath.Join(root, "mobilelab", "scenarios", "broken.yaml")
	if err := os.WriteFile(broken, []byte("name: broken\nunknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root); err == nil || !strings.Contains(err.Error(), "broken.yaml") {
		t.Fatalf("expected scenario preflight failure, got %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, config.DefaultFilename))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "schema_version") {
		t.Fatal("failed preflight modified the configuration")
	}
}

func TestPlanAllowsConfigurationOnlyProject(t *testing.T) {
	root := t.TempDir()
	if err := config.Write(filepath.Join(root, config.DefaultFilename), config.Default("current")); err != nil {
		t.Fatal(err)
	}
	plan, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.ChangeCount() != 0 || len(plan.Documents) != 1 {
		t.Fatalf("unexpected configuration-only plan: %#v", plan)
	}
}

func legacyProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	cfg := config.Default("legacy")
	cfg.SchemaVersion = 0
	configPath := filepath.Join(root, config.DefaultFilename)
	if err := config.Write(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.TrimPrefix(string(data), "schema_version: 1\n"))
	if err := os.WriteFile(configPath, data, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(configPath, 0o640); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, "mobilelab", "scenarios")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "smoke.yaml"), []byte("# preserved comment\nname: Smoke\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}
