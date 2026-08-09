package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var output bytes.Buffer
	runner := New(&output, &output, t.TempDir())
	if err := runner.Run(context.Background(), []string{"version"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), Version) {
		t.Fatalf("output does not contain version: %q", output.String())
	}
}

func TestUnknownCommandIsActionable(t *testing.T) {
	runner := New(&bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	err := runner.Run(context.Background(), []string{"wat"})
	if err == nil || !strings.Contains(err.Error(), "mobilelab help") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScenarioListParsesGeneratedScenarios(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "mobilelab", "scenarios")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "payment.yaml"), []byte("name: Payment failure\nexpect:\n  - response: {status: 500}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	runner := New(&output, &output, root)
	if err := runner.Run(context.Background(), []string{"scenario", "list"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "payment") || !strings.Contains(output.String(), "Payment failure") {
		t.Fatalf("unexpected scenario list: %q", output.String())
	}
}
