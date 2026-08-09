package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func validRequest() string {
	return `{"protocol":"mobilelab.plugin/v1","request_id":"request-1","action":"echo","input":{"name":"MobileLab"},"context":{"project_dir":"/project","plugin_dir":"/project/mobilelab/plugins/echo","mobilelab_version":"0.7.0"}}`
}

func TestServeExecutesTypedHandlerAndCorrelatesResponse(t *testing.T) {
	var output bytes.Buffer
	err := Serve(context.Background(), strings.NewReader(validRequest()), &output, HandlerFunc(
		func(_ context.Context, action string, input json.RawMessage, invocation InvocationContext) (Result, error) {
			if action != "echo" || invocation.MobileLabVersion != "0.7.0" || !strings.Contains(string(input), "MobileLab") {
				t.Fatalf("unexpected invocation: %s %s %#v", action, input, invocation)
			}
			return Result{Message: "done", Output: map[string]bool{"echoed": true}}, nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	response, err := DecodeResponse(&output)
	if err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.RequestID != "request-1" || response.Message != "done" || !strings.Contains(string(response.Output), "echoed") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestServeTurnsHandlerFailureIntoProtocolFailure(t *testing.T) {
	var output bytes.Buffer
	err := Serve(context.Background(), strings.NewReader(validRequest()), &output, HandlerFunc(
		func(context.Context, string, json.RawMessage, InvocationContext) (Result, error) {
			return Result{}, errors.New("fixture is invalid")
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	response, err := DecodeResponse(&output)
	if err != nil {
		t.Fatal(err)
	}
	if response.Success || response.Error == nil || response.Error.Code != "execution_failed" || response.Error.Message != "fixture is invalid" {
		t.Fatalf("unexpected failure response: %#v", response)
	}
}

func TestDecodeRequestRejectsUnknownFieldsAndOversizedMessages(t *testing.T) {
	unknown := strings.Replace(validRequest(), `"action":"echo"`, `"action":"echo","secret":true`, 1)
	if _, err := DecodeRequest(strings.NewReader(unknown)); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected strict field error, got %v", err)
	}
	if _, err := DecodeRequest(strings.NewReader(strings.Repeat("x", MaxMessageBytes+1))); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size error, got %v", err)
	}
}

func TestEncodeResponseRejectsOversizedMessages(t *testing.T) {
	response := Response{
		Protocol: ProtocolVersion, RequestID: "request-1", Success: true,
		Output: json.RawMessage(`"` + strings.Repeat("x", MaxMessageBytes) + `"`),
	}
	if err := EncodeResponse(&bytes.Buffer{}, response); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected encoded size error, got %v", err)
	}
}
