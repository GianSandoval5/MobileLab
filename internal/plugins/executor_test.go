package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	protocol "github.com/mobilelab-dev/mobilelab/pkg/plugin"
)

type processFunc func(context.Context, string, string, []string, []byte, int) ([]byte, error)

func (function processFunc) Run(ctx context.Context, executable, directory string, environment []string, input []byte, limit int) ([]byte, error) {
	return function(ctx, executable, directory, environment, input, limit)
}

func testDescriptor() Descriptor {
	return Descriptor{
		Manifest:   Manifest{Name: "echo", Actions: []ActionDefinition{{Name: "echo"}}},
		ProjectDir: "/project", Directory: "/project/mobilelab/plugins/echo", Executable: "/project/mobilelab/plugins/echo/plugin",
	}
}

func TestExecutorCorrelatesResponseAndDoesNotInheritSecrets(t *testing.T) {
	t.Setenv("MOBILELAB_TEST_SECRET", "must-not-leak")
	process := processFunc(func(_ context.Context, executable, directory string, environment []string, input []byte, limit int) ([]byte, error) {
		if strings.Contains(strings.Join(environment, "\n"), "MOBILELAB_TEST_SECRET") {
			t.Fatalf("secret environment variable leaked: %v", environment)
		}
		if executable == "" || directory == "" || limit != protocol.MaxMessageBytes {
			t.Fatalf("unexpected process invocation: %q %q %d", executable, directory, limit)
		}
		request, err := protocol.DecodeRequest(bytes.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if request.RequestID != "fixed-request" || request.Action != "echo" || request.Context.ProjectDir != "/project" {
			t.Fatalf("unexpected request: %#v", request)
		}
		var output bytes.Buffer
		err = protocol.EncodeResponse(&output, protocol.Response{
			Protocol: protocol.ProtocolVersion, RequestID: request.RequestID, Success: true, Message: "echoed", Output: json.RawMessage(`{"ok":true}`),
		})
		return output.Bytes(), err
	})
	executor := Executor{Process: process, RequestID: func() (string, error) { return "fixed-request", nil }}
	response, err := executor.Execute(context.Background(), testDescriptor(), "echo", json.RawMessage(`{"name":"MobileLab"}`), time.Second, "0.7.0")
	if err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.Message != "echoed" || !strings.Contains(string(response.Output), "true") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestExecutorRejectsUndeclaredActionsAndMismatchedResponses(t *testing.T) {
	executor := Executor{}
	if _, err := executor.Execute(context.Background(), testDescriptor(), "delete-everything", nil, time.Second, "0.7.0"); err == nil || !strings.Contains(err.Error(), "does not declare") {
		t.Fatalf("expected undeclared action error, got %v", err)
	}
	process := processFunc(func(context.Context, string, string, []string, []byte, int) ([]byte, error) {
		var output bytes.Buffer
		err := protocol.EncodeResponse(&output, protocol.Response{Protocol: protocol.ProtocolVersion, RequestID: "another-request", Success: true})
		return output.Bytes(), err
	})
	executor = Executor{Process: process, RequestID: func() (string, error) { return "fixed-request", nil }}
	if _, err := executor.Execute(context.Background(), testDescriptor(), "echo", nil, time.Second, "0.7.0"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected correlation error, got %v", err)
	}
}

func TestExecutorBoundsOutputAndTimeout(t *testing.T) {
	oversized := processFunc(func(context.Context, string, string, []string, []byte, int) ([]byte, error) {
		return []byte(strings.Repeat("x", protocol.MaxMessageBytes+1)), nil
	})
	executor := Executor{Process: oversized, RequestID: func() (string, error) { return "fixed-request", nil }}
	if _, err := executor.Execute(context.Background(), testDescriptor(), "echo", nil, time.Second, "0.7.0"); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected bounded output error, got %v", err)
	}
	blocking := processFunc(func(ctx context.Context, _ string, _ string, _ []string, _ []byte, _ int) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	executor = Executor{Process: blocking, RequestID: func() (string, error) { return "fixed-request", nil }}
	if _, err := executor.Execute(context.Background(), testDescriptor(), "echo", nil, time.Millisecond, "0.7.0"); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestExecutorReportsProtocolFailureWithoutRunningShell(t *testing.T) {
	process := processFunc(func(context.Context, string, string, []string, []byte, int) ([]byte, error) {
		return nil, errors.New("exit status 2")
	})
	executor := Executor{Process: process, RequestID: func() (string, error) { return "fixed-request", nil }}
	if _, err := executor.Execute(context.Background(), testDescriptor(), "echo", nil, time.Second, "0.7.0"); err == nil || !strings.Contains(err.Error(), `run plugin "echo" action "echo": exit status 2`) {
		t.Fatalf("expected process error, got %v", err)
	}
}
