package domain

import (
	"context"
	"time"
)

const AppEventProtocolVersion = 1

type AppFramework string

const (
	FrameworkFlutter     AppFramework = "flutter"
	FrameworkReactNative AppFramework = "react-native"
)

type AppEventKind string

const (
	AppEventLifecycle AppEventKind = "lifecycle"
	AppEventMarker    AppEventKind = "marker"
	AppEventAssertion AppEventKind = "assertion"
)

type AppEvent struct {
	ProtocolVersion int            `json:"protocolVersion"`
	Framework       AppFramework   `json:"framework"`
	Kind            AppEventKind   `json:"kind"`
	Name            string         `json:"name"`
	Passed          *bool          `json:"passed,omitempty"`
	SessionID       string         `json:"sessionId,omitempty"`
	Attributes      map[string]any `json:"attributes,omitempty"`
	Timestamp       time.Time      `json:"timestamp"`
}

type AppEventRepository interface {
	SaveAppEvent(context.Context, AppEvent) error
	RecentAppEvents(context.Context, int) ([]AppEvent, error)
}
