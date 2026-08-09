package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
