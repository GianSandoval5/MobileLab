package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/schemas"
)

func TestLoadStrictValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, DefaultFilename)
	contents := `
project:
  name: sample
server:
  host: 127.0.0.1
  port: 4566
dashboard:
  enabled: true
sandbox:
  latency: 25
  error_rate: 0
auth:
  enabled: true
device:
  auto_detect: false
endpoints:
  - path: /api/users
    method: GET
    response:
      status: 200
      body:
        users: []
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Project.Name != "sample" || len(cfg.Endpoints) != 1 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestConfigSchemaVersionCompatibilityAndWrite(t *testing.T) {
	legacy := Default("legacy")
	legacy.SchemaVersion = 0
	path := filepath.Join(t.TempDir(), DefaultFilename)
	if err := Write(path, legacy); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "schema_version: 1\n") {
		t.Fatalf("write did not emit stable schema version: %s", data)
	}
	cfg, err := Load(path)
	if err != nil || cfg.SchemaVersion != schemas.ConfigVersion {
		t.Fatalf("unexpected versioned config=%#v err=%v", cfg, err)
	}

	newer := strings.Replace(string(data), "schema_version: 1", "schema_version: 2", 1)
	if err := os.WriteFile(path, []byte(newer), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "newer than supported") {
		t.Fatalf("expected newer schema rejection, got %v", err)
	}
}

func TestLoadAcceptsLegacyConfigAndRejectsMultipleOrOversizedDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultFilename)
	legacy := "project: {name: legacy}\nserver: {host: 127.0.0.1, port: 4566}\nendpoints: []\n"
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil || cfg.SchemaVersion != 0 {
		t.Fatalf("legacy config was not accepted: %#v %v", cfg, err)
	}
	if err := os.WriteFile(path, []byte(legacy+"---\nproject: {name: second}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "exactly one YAML document") {
		t.Fatalf("expected multiple document rejection, got %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", schemas.MaxYAMLBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size rejection, got %v", err)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, DefaultFilename)
	contents := `
project: {name: sample}
server: {host: 127.0.0.1, port: 4566, mystery: true}
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "mystery") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestValidateCollectsActionableErrors(t *testing.T) {
	cfg := Default("sample")
	cfg.Server.Port = 70000
	cfg.Endpoints[0].Path = "api/profile"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	for _, want := range []string{"server.port", "must begin with /"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
}

func TestValidatePushFixtures(t *testing.T) {
	cfg := Default("sample")
	cfg.Push["empty"] = PushFixture{}
	cfg.Push["reserved"] = PushFixture{Body: "Hello", Data: map[string]any{"aps": "override"}}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "push.empty") || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("unexpected push validation error: %v", err)
	}
}
