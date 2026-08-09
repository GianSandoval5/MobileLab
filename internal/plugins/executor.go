package plugins

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	protocol "github.com/mobilelab-dev/mobilelab/pkg/plugin"
)

const DefaultTimeout = 30 * time.Second

type ProcessRunner interface {
	Run(context.Context, string, string, []string, []byte, int) ([]byte, error)
}

type ExecProcessRunner struct{}

func (ExecProcessRunner) Run(ctx context.Context, executable, directory string, environment []string, input []byte, limit int) ([]byte, error) {
	command := exec.CommandContext(ctx, executable)
	command.Dir = directory
	command.Env = environment
	command.Stdin = bytes.NewReader(input)
	stdout := &cappedBuffer{limit: limit}
	stderr := &cappedBuffer{limit: 4096}
	command.Stdout = stdout
	command.Stderr = stderr
	err := command.Run()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if stdout.exceeded {
		return nil, fmt.Errorf("plugin output exceeds %d bytes", limit)
	}
	if err != nil {
		return nil, fmt.Errorf("plugin process failed: %w", err)
	}
	return append([]byte(nil), stdout.Bytes()...), nil
}

type Executor struct {
	Process   ProcessRunner
	RequestID func() (string, error)
}

func (executor Executor) Execute(ctx context.Context, descriptor Descriptor, action string, input json.RawMessage, timeout time.Duration, mobileLabVersion string) (protocol.Response, error) {
	if !descriptor.Manifest.Supports(action) {
		return protocol.Response{}, fmt.Errorf("plugin %q does not declare action %q", descriptor.Manifest.Name, action)
	}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if len(input) > protocol.MaxMessageBytes || !json.Valid(input) {
		return protocol.Response{}, fmt.Errorf("plugin input must be valid JSON no larger than %d bytes", protocol.MaxMessageBytes)
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	requestID := executor.RequestID
	if requestID == nil {
		requestID = randomRequestID
	}
	id, err := requestID()
	if err != nil {
		return protocol.Response{}, fmt.Errorf("create plugin request ID: %w", err)
	}
	request := protocol.Request{
		Protocol: protocol.ProtocolVersion, RequestID: id, Action: action, Input: input,
		Context: protocol.InvocationContext{ProjectDir: descriptor.ProjectDir, PluginDir: descriptor.Directory, MobileLabVersion: mobileLabVersion},
	}
	if err := request.Validate(); err != nil {
		return protocol.Response{}, err
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("encode plugin request: %w", err)
	}
	if len(encoded) > protocol.MaxMessageBytes {
		return protocol.Response{}, fmt.Errorf("encoded plugin request exceeds %d bytes", protocol.MaxMessageBytes)
	}
	process := executor.Process
	if process == nil {
		process = ExecProcessRunner{}
	}
	runContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	output, err := process.Run(runContext, descriptor.Executable, descriptor.Directory, minimalEnvironment(descriptor.Manifest.Name), encoded, protocol.MaxMessageBytes)
	if err != nil {
		if runContext.Err() == context.DeadlineExceeded {
			return protocol.Response{}, fmt.Errorf("plugin %q timed out after %s", descriptor.Manifest.Name, timeout)
		}
		return protocol.Response{}, fmt.Errorf("run plugin %q action %q: %w", descriptor.Manifest.Name, action, err)
	}
	response, err := protocol.DecodeResponse(bytes.NewReader(output))
	if err != nil {
		return protocol.Response{}, fmt.Errorf("plugin %q returned an invalid response: %w", descriptor.Manifest.Name, err)
	}
	if response.RequestID != id {
		return protocol.Response{}, fmt.Errorf("plugin %q response request_id does not match the invocation", descriptor.Manifest.Name)
	}
	if !response.Success {
		return protocol.Response{}, fmt.Errorf("plugin %q action %q failed (%s): %s", descriptor.Manifest.Name, action, response.Error.Code, response.Error.Message)
	}
	return response, nil
}

func randomRequestID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func minimalEnvironment(pluginName string) []string {
	names := []string{"PATH", "TMPDIR", "TEMP", "TMP", "SYSTEMROOT", "SystemRoot", "COMSPEC", "ComSpec", "PATHEXT"}
	environment := make([]string, 0, len(names)+2)
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		key := name
		if runtime.GOOS == "windows" {
			key = strings.ToLower(name)
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		if value, found := os.LookupEnv(name); found && value != "" {
			environment = append(environment, name+"="+value)
		}
	}
	environment = append(environment, "MOBILELAB_PLUGIN_PROTOCOL="+protocol.ProtocolVersion, "MOBILELAB_PLUGIN_NAME="+pluginName)
	return environment
}

type cappedBuffer struct {
	bytes.Buffer
	limit    int
	exceeded bool
}

func (buffer *cappedBuffer) Write(data []byte) (int, error) {
	originalLength := len(data)
	remaining := buffer.limit - buffer.Len()
	if remaining <= 0 {
		buffer.exceeded = buffer.exceeded || originalLength > 0
		return originalLength, nil
	}
	if len(data) > remaining {
		data = data[:remaining]
		buffer.exceeded = true
	}
	_, _ = buffer.Buffer.Write(data)
	return originalLength, nil
}
