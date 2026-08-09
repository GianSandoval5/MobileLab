package device

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type ProcessRunner interface {
	LookPath(string) (string, error)
	Run(context.Context, string, ...string) ([]byte, error)
	Start(context.Context, string, ...string) error
}

type ExecRunner struct{}

func (ExecRunner) LookPath(name string) (string, error) {
	resolved, pathErr := exec.LookPath(name)
	if pathErr == nil {
		return resolved, nil
	}
	relative := map[string]string{
		"adb":      filepath.Join("platform-tools", executableName("adb")),
		"emulator": filepath.Join("emulator", executableName("emulator")),
	}[name]
	if relative != "" {
		for _, variable := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
			root := strings.TrimSpace(os.Getenv(variable))
			if root == "" {
				continue
			}
			candidate := filepath.Join(root, relative)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	return "", pathErr
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w: %s", name, err, output)
	}
	return output, nil
}

func (ExecRunner) Start(ctx context.Context, name string, args ...string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// A started device must outlive the short-lived CLI command that requested it.
	command := exec.Command(name, args...)
	if err := command.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}
	go func() {
		_ = command.Wait()
	}()
	return nil
}
