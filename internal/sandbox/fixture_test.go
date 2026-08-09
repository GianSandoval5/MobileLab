package sandbox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFixtureSubstitutesVariables(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "user.json"), []byte(`{"id":"{{userId}}"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := loadFixture(root, "user.json", map[string]string{"userId": "123"})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"id":"123"}` {
		t.Fatalf("unexpected fixture: %s", data)
	}
}

func TestRenderVariablesPreservesJSONEscaping(t *testing.T) {
	value := renderVariables(map[string]any{"message": "Hello {{name}}"}, map[string]string{"name": `A \"quoted\" user`}).(map[string]any)
	if value["message"] != `Hello A \"quoted\" user` {
		t.Fatalf("unexpected rendered value: %#v", value)
	}
}

func TestLoadFixtureRejectsPathTraversal(t *testing.T) {
	_, err := loadFixture(t.TempDir(), "../secret.json", nil)
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected traversal error, got %v", err)
	}
}

func TestLoadFixtureRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(outside, []byte(`{"secret":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked.json")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err := loadFixture(root, "linked.json", nil)
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected symlink escape error, got %v", err)
	}
}
