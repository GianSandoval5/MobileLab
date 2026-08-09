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
	FrameworkAndroid     AppFramework = "android"
	FrameworkIOS         AppFramework = "ios"
	FrameworkCapacitor   AppFramework = "capacitor"
	AppFrameworkChoices               = "flutter, react-native, android, ios, or capacitor"
)

func (framework AppFramework) Valid() bool {
	switch framework {
	case FrameworkFlutter, FrameworkReactNative, FrameworkAndroid, FrameworkIOS, FrameworkCapacitor:
		return true
	default:
		return false
	}
}

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
