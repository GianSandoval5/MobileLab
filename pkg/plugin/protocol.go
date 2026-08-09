package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

const (
	ProtocolVersion = "mobilelab.plugin/v1"
	MaxMessageBytes = 1 << 20
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

type Request struct {
	Protocol  string            `json:"protocol"`
	RequestID string            `json:"request_id"`
	Action    string            `json:"action"`
	Input     json.RawMessage   `json:"input,omitempty"`
	Context   InvocationContext `json:"context"`
}

type InvocationContext struct {
	ProjectDir       string `json:"project_dir"`
	PluginDir        string `json:"plugin_dir"`
	MobileLabVersion string `json:"mobilelab_version"`
}

type Response struct {
	Protocol  string          `json:"protocol"`
	RequestID string          `json:"request_id"`
	Success   bool            `json:"success"`
	Message   string          `json:"message,omitempty"`
	Output    json.RawMessage `json:"output,omitempty"`
	Error     *ResponseError  `json:"error,omitempty"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Result struct {
	Message string
	Output  any
}

type Handler interface {
	Execute(context.Context, string, json.RawMessage, InvocationContext) (Result, error)
}

type HandlerFunc func(context.Context, string, json.RawMessage, InvocationContext) (Result, error)

func (function HandlerFunc) Execute(ctx context.Context, action string, input json.RawMessage, invocation InvocationContext) (Result, error) {
	return function(ctx, action, input, invocation)
}

func Serve(ctx context.Context, input io.Reader, output io.Writer, handler Handler) error {
	if handler == nil {
		return fmt.Errorf("plugin handler is required")
	}
	request, err := DecodeRequest(input)
	if err != nil {
		return err
	}
	result, executionErr := handler.Execute(ctx, request.Action, request.Input, request.Context)
	response := Response{Protocol: ProtocolVersion, RequestID: request.RequestID, Success: executionErr == nil, Message: result.Message}
	if executionErr != nil {
		response.Error = &ResponseError{Code: "execution_failed", Message: executionErr.Error()}
	} else if result.Output != nil {
		encoded, err := json.Marshal(result.Output)
		if err != nil {
			return fmt.Errorf("encode plugin output: %w", err)
		}
		response.Output = encoded
	}
	return EncodeResponse(output, response)
}

func DecodeRequest(reader io.Reader) (Request, error) {
	var request Request
	if err := decodeStrictMessage(reader, &request); err != nil {
		return Request{}, fmt.Errorf("decode plugin request: %w", err)
	}
	if err := request.Validate(); err != nil {
		return Request{}, fmt.Errorf("invalid plugin request: %w", err)
	}
	return request, nil
}

func DecodeResponse(reader io.Reader) (Response, error) {
	var response Response
	if err := decodeStrictMessage(reader, &response); err != nil {
		return Response{}, fmt.Errorf("decode plugin response: %w", err)
	}
	if err := response.Validate(); err != nil {
		return Response{}, fmt.Errorf("invalid plugin response: %w", err)
	}
	return response, nil
}

func EncodeResponse(writer io.Writer, response Response) error {
	if err := response.Validate(); err != nil {
		return fmt.Errorf("invalid plugin response: %w", err)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode plugin response: %w", err)
	}
	if len(encoded)+1 > MaxMessageBytes {
		return fmt.Errorf("encode plugin response: message exceeds %d bytes", MaxMessageBytes)
	}
	encoded = append(encoded, '\n')
	if _, err := writer.Write(encoded); err != nil {
		return fmt.Errorf("encode plugin response: %w", err)
	}
	return nil
}

func (request Request) Validate() error {
	if request.Protocol != ProtocolVersion {
		return fmt.Errorf("protocol must be %q", ProtocolVersion)
	}
	if !identifierPattern.MatchString(request.RequestID) {
		return fmt.Errorf("request_id is invalid")
	}
	if !identifierPattern.MatchString(request.Action) {
		return fmt.Errorf("action is invalid")
	}
	if len(request.Input) > 0 && !json.Valid(request.Input) {
		return fmt.Errorf("input must be valid JSON")
	}
	if request.Context.ProjectDir == "" || request.Context.PluginDir == "" || request.Context.MobileLabVersion == "" {
		return fmt.Errorf("context requires project_dir, plugin_dir, and mobilelab_version")
	}
	return nil
}

func (response Response) Validate() error {
	if response.Protocol != ProtocolVersion {
		return fmt.Errorf("protocol must be %q", ProtocolVersion)
	}
	if !identifierPattern.MatchString(response.RequestID) {
		return fmt.Errorf("request_id is invalid")
	}
	if response.Success && response.Error != nil {
		return fmt.Errorf("successful response cannot contain error")
	}
	if !response.Success {
		if response.Error == nil || !identifierPattern.MatchString(response.Error.Code) || response.Error.Message == "" {
			return fmt.Errorf("failed response requires a valid error code and message")
		}
	}
	if len(response.Output) > 0 {
		if len(response.Output) > MaxMessageBytes {
			return fmt.Errorf("output exceeds %d bytes", MaxMessageBytes)
		}
		if !json.Valid(response.Output) {
			return fmt.Errorf("output must be valid JSON")
		}
	}
	return nil
}

func decodeStrictMessage(reader io.Reader, target any) error {
	data, err := io.ReadAll(io.LimitReader(reader, MaxMessageBytes+1))
	if err != nil {
		return err
	}
	if len(data) > MaxMessageBytes {
		return fmt.Errorf("message exceeds %d bytes", MaxMessageBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("message must contain exactly one JSON value")
		}
		return err
	}
	return nil
}
