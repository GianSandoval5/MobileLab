package device

import (
	"context"
	"fmt"
	"os/exec"
)

type ProcessRunner interface {
	LookPath(string) (string, error)
	Run(context.Context, string, ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w: %s", name, err, output)
	}
	return output, nil
}
