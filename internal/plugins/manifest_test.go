package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const validManifest = `api_version: mobilelab.plugin/v1
name: echo
version: 1.2.3
description: Echoes structured input
executable: mobilelab-plugin-echo
actions:
  - name: echo
    description: Return the provided JSON
`

func TestLoadManifestIsStrictAndValidatesSemanticVersioning(t *testing.T) {
	path := filepath.Join(t.TempDir(), ManifestFilename)
	if err := os.WriteFile(path, []byte(validManifest), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := LoadManifest(path)
	if err != nil || !manifest.Supports("echo") {
		t.Fatalf("unexpected manifest=%#v err=%v", manifest, err)
	}
	unknown := strings.Replace(validManifest, "description: Echoes structured input", "description: Echoes structured input\nunknown: true", 1)
	if err := os.WriteFile(path, []byte(unknown), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
	invalidVersion := strings.Replace(validManifest, "version: 1.2.3", "version: latest", 1)
	if err := os.WriteFile(path, []byte(invalidVersion), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), "Semantic Versioning") {
		t.Fatalf("expected version error, got %v", err)
	}
	leadingZero := strings.Replace(validManifest, "version: 1.2.3", "version: 1.2.3-01", 1)
	if err := os.WriteFile(path, []byte(leadingZero), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), "Semantic Versioning") {
		t.Fatalf("expected numeric prerelease error, got %v", err)
	}
	multiple := validManifest + "---\nname: another\n"
	if err := os.WriteFile(path, []byte(multiple), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), "exactly one YAML document") {
		t.Fatalf("expected multiple document error, got %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", MaxManifestBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected manifest size error, got %v", err)
	}
}

func TestCatalogDiscoversValidPluginsAndReportsInvalidOnes(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "echo", validManifest)
	brokenDirectory := filepath.Join(root, "mobilelab", "plugins", "broken")
	if err := os.MkdirAll(brokenDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brokenDirectory, ManifestFilename), []byte("api_version: wrong\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	descriptors, issues, err := (Catalog{ProjectDir: root}).Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(descriptors) != 1 || descriptors[0].Manifest.Name != "echo" || len(descriptors[0].SHA256) != 64 {
		t.Fatalf("unexpected descriptors: %#v", descriptors)
	}
	if len(issues) != 1 || !strings.Contains(issues[0].Path, "broken") {
		t.Fatalf("unexpected issues: %#v", issues)
	}
}

func TestCatalogRejectsExecutableSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not generally available to unprivileged Windows CI")
	}
	root := t.TempDir()
	directory := filepath.Join(root, "mobilelab", "plugins", "echo")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, ManifestFilename), []byte(validManifest), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside")
	if err := os.WriteFile(outside, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, "mobilelab-plugin-echo")); err != nil {
		t.Fatal(err)
	}
	if _, err := (Catalog{ProjectDir: root}).Load("echo"); err == nil || !strings.Contains(err.Error(), "must remain inside") {
		t.Fatalf("expected confinement error, got %v", err)
	}
}

func writeTestPlugin(t *testing.T, root, name, manifest string) {
	t.Helper()
	directory := filepath.Join(root, "mobilelab", "plugins", name)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, ManifestFilename), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(directory, "mobilelab-plugin-echo")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	if err := os.WriteFile(executable, []byte("test executable"), 0o755); err != nil {
		t.Fatal(err)
	}
}
