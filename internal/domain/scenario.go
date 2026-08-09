package domain

import (
	"context"
	"fmt"
	"time"
)

type ScenarioDefinition struct {
	Name       string
	Backend    ScenarioBackend
	Auth       ScenarioAuth
	Device     ScenarioDevice
	Steps      []ScenarioStep
	Assertions []ScenarioAssertion
}

type ScenarioBackend struct {
	LatencyMS int
	Error     int
}

type ScenarioAuth struct {
	Token string
}

type ScenarioDevice struct {
	Network NetworkCondition
}

type ScenarioStepKind string

const (
	StepLaunchApp    ScenarioStepKind = "launch_app"
	StepStopApp      ScenarioStepKind = "stop_app"
	StepOpenDeepLink ScenarioStepKind = "open_deeplink"
)

type ScenarioStep struct {
	Kind  ScenarioStepKind
	Value string
}

type ScenarioAssertion struct {
	Request  *RequestExpectation
	Response *ResponseExpectation
	AppEvent *AppEventExpectation
}

type AppEventExpectation struct {
	Framework AppFramework
	Kind      AppEventKind
	Name      string
	Passed    *bool
}

type RequestExpectation struct {
	Method string
	Path   string
}

type ResponseExpectation struct {
	Status int
}

type ScenarioRunOptions struct {
	DeviceID string
	AppID    string
	Timeout  time.Duration
}

type ScenarioResult struct {
	Name       string          `json:"name"`
	Passed     bool            `json:"passed"`
	StartedAt  time.Time       `json:"started_at"`
	DurationMS int64           `json:"duration_ms"`
	Steps      []ScenarioCheck `json:"steps"`
	Assertions []ScenarioCheck `json:"assertions"`
	Error      string          `json:"error,omitempty"`
}

type ScenarioCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

type ScenarioParser interface {
	Parse([]byte) (ScenarioDefinition, error)
}

type ScenarioEnvironment interface {
	SetLatency(context.Context, int) error
	SetError(context.Context, int) error
	SetAuthExpired(context.Context, bool) error
	Reset(context.Context) error
	RecentRequests(context.Context, int) ([]RequestRecord, error)
	RecentAppEvents(context.Context, int) ([]AppEvent, error)
}

type ScenarioRunRepository interface {
	Save(context.Context, ScenarioResult) error
	Recent(context.Context, int) ([]ScenarioResult, error)
}

func (d ScenarioDefinition) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("scenario name is required")
	}
	if d.Backend.LatencyMS < 0 {
		return fmt.Errorf("backend latency cannot be negative")
	}
	if d.Backend.Error != 0 && (d.Backend.Error < 400 || d.Backend.Error > 599) {
		return fmt.Errorf("backend error must be between 400 and 599")
	}
	if d.Auth.Token != "" && d.Auth.Token != "valid" && d.Auth.Token != "expired" {
		return fmt.Errorf("auth token must be valid or expired")
	}
	if d.Device.Network != "" && d.Device.Network != NetworkOnline && d.Device.Network != NetworkOffline && d.Device.Network != NetworkSlow {
		return fmt.Errorf("device network must be online, offline, or slow")
	}
	for index, step := range d.Steps {
		switch step.Kind {
		case StepLaunchApp, StepStopApp:
		case StepOpenDeepLink:
			if step.Value == "" {
				return fmt.Errorf("steps[%d] open_deeplink requires a value", index)
			}
		default:
			return fmt.Errorf("steps[%d] has unsupported kind %q", index, step.Kind)
		}
	}
	for index, assertion := range d.Assertions {
		defined := 0
		if assertion.Request != nil {
			defined++
		}
		if assertion.Response != nil {
			defined++
		}
		if assertion.AppEvent != nil {
			defined++
		}
		if defined != 1 {
			return fmt.Errorf("assertions[%d] must define exactly one request, response, or app_event", index)
		}
		if assertion.Request != nil && (assertion.Request.Method == "" || assertion.Request.Path == "") {
			return fmt.Errorf("assertions[%d].request requires method and path", index)
		}
		if assertion.Response != nil && (assertion.Response.Status < 100 || assertion.Response.Status > 599) {
			return fmt.Errorf("assertions[%d].response.status must be between 100 and 599", index)
		}
		if expectation := assertion.AppEvent; expectation != nil {
			if expectation.Name == "" {
				return fmt.Errorf("assertions[%d].app_event requires name", index)
			}
			switch expectation.Kind {
			case AppEventLifecycle, AppEventMarker, AppEventAssertion:
			default:
				return fmt.Errorf("assertions[%d].app_event kind must be lifecycle, marker, or assertion", index)
			}
			if expectation.Framework != "" && expectation.Framework != FrameworkFlutter && expectation.Framework != FrameworkReactNative {
				return fmt.Errorf("assertions[%d].app_event framework must be flutter or react-native", index)
			}
			if expectation.Kind != AppEventAssertion && expectation.Passed != nil {
				return fmt.Errorf("assertions[%d].app_event passed is valid only for assertion events", index)
			}
		}
	}
	return nil
}
