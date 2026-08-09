package device

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecRunnerFindsAndroidToolsFromSDKRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PATH", "")
	t.Setenv("ANDROID_HOME", root)
	t.Setenv("ANDROID_SDK_ROOT", "")

	want := filepath.Join(root, "emulator", executableName("emulator"))
	if err := os.MkdirAll(filepath.Dir(want), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(want, nil, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := (ExecRunner{}).LookPath("emulator")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
