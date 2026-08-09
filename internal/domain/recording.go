package domain

import (
	"context"
	"fmt"
	"time"
)

type CaptureKind string

const (
	CaptureHTTPExchange CaptureKind = "http_exchange"
	CaptureEnvironment  CaptureKind = "environment"
	CaptureDeepLink     CaptureKind = "deeplink"
)

type EnvironmentState struct {
	LatencyMS   int  `json:"latency_ms"`
	ForcedError int  `json:"forced_error,omitempty"`
	AuthExpired bool `json:"auth_expired"`
}

type EnvironmentMutation struct {
	Action      string `json:"action"`
	LatencyMS   int    `json:"latency_ms,omitempty"`
	ForcedError int    `json:"forced_error,omitempty"`
	AuthExpired bool   `json:"auth_expired,omitempty"`
}

type DeepLinkCapture struct {
	URL      string `json:"url"`
	Platform string `json:"platform,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
}

type CaptureEvent struct {
	Sequence    int64                `json:"sequence"`
	Kind        CaptureKind          `json:"kind"`
	Timestamp   time.Time            `json:"timestamp"`
	HTTP        *RequestRecord       `json:"http,omitempty"`
	Environment *EnvironmentMutation `json:"environment,omitempty"`
	DeepLink    *DeepLinkCapture     `json:"deeplink,omitempty"`
}

func (event CaptureEvent) Validate() error {
	defined := 0
	if event.HTTP != nil {
		defined++
	}
	if event.Environment != nil {
		defined++
	}
	if event.DeepLink != nil {
		defined++
	}
	if defined != 1 {
		return fmt.Errorf("capture event must define exactly one payload")
	}
	switch event.Kind {
	case CaptureHTTPExchange:
		if event.HTTP == nil || event.HTTP.Method == "" || event.HTTP.Path == "" {
			return fmt.Errorf("HTTP capture requires method and path")
		}
	case CaptureEnvironment:
		if event.Environment == nil || event.Environment.Action == "" {
			return fmt.Errorf("environment capture requires an action")
		}
	case CaptureDeepLink:
		if event.DeepLink == nil || event.DeepLink.URL == "" {
			return fmt.Errorf("deep-link capture requires a URL")
		}
	default:
		return fmt.Errorf("unsupported capture kind %q", event.Kind)
	}
	return nil
}

type Recording struct {
	Name               string           `json:"name"`
	StartedAt          time.Time        `json:"started_at"`
	StoppedAt          time.Time        `json:"stopped_at,omitempty"`
	InitialEnvironment EnvironmentState `json:"initial_environment"`
	Events             []CaptureEvent   `json:"events"`
}

type CaptureSink interface {
	RecordCapture(context.Context, CaptureEvent) error
}

type RecordingController interface {
	StartRecording(context.Context, string, EnvironmentState) (Recording, error)
	StopRecording(context.Context) (Recording, error)
	ActiveRecording(context.Context) (Recording, bool)
}
