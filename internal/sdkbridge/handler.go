package sdkbridge

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/sandbox"
)

const MaxEventBytes = 64 << 10

var eventNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type Handler struct {
	Repository domain.AppEventRepository
	Events     domain.EventPublisher
	Now        func() time.Time
}

type eventInput struct {
	ProtocolVersion int                 `json:"protocolVersion"`
	Framework       domain.AppFramework `json:"framework"`
	Kind            domain.AppEventKind `json:"kind"`
	Name            string              `json:"name"`
	Passed          *bool               `json:"passed,omitempty"`
	SessionID       string              `json:"sessionId,omitempty"`
	Attributes      map[string]any      `json:"attributes,omitempty"`
}

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.Repository == nil {
		http.Error(writer, "SDK bridge unavailable", http.StatusServiceUnavailable)
		return
	}
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(writer, "content type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, MaxEventBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var input eventInput
	if err := decoder.Decode(&input); err != nil {
		http.Error(writer, "invalid SDK event: "+err.Error(), http.StatusBadRequest)
		return
	}
	event := domain.AppEvent{
		ProtocolVersion: input.ProtocolVersion, Framework: input.Framework, Kind: input.Kind,
		Name: input.Name, Passed: input.Passed, SessionID: input.SessionID, Attributes: input.Attributes,
	}
	if err := ensureJSONEnd(decoder); err != nil {
		http.Error(writer, "invalid SDK event: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validate(&event); err != nil {
		http.Error(writer, "invalid SDK event: "+err.Error(), http.StatusBadRequest)
		return
	}

	event.Attributes = sanitizedAttributes(event.Attributes)
	now := h.Now
	if now == nil {
		now = time.Now
	}
	event.Timestamp = now().UTC()
	if err := h.Repository.SaveAppEvent(request.Context(), event); err != nil {
		http.Error(writer, "unable to store SDK event", http.StatusInternalServerError)
		return
	}
	if h.Events != nil {
		_ = h.Events.Publish(request.Context(), domain.Event{
			Type: domain.EventAppReported, Version: 1, Timestamp: event.Timestamp, Payload: event,
		})
	}
	writer.WriteHeader(http.StatusAccepted)
}

func validate(event *domain.AppEvent) error {
	if event.ProtocolVersion != domain.AppEventProtocolVersion {
		return fmt.Errorf("protocolVersion must be %d", domain.AppEventProtocolVersion)
	}
	switch event.Framework {
	case domain.FrameworkFlutter, domain.FrameworkReactNative:
	default:
		return fmt.Errorf("framework must be flutter or react-native")
	}
	switch event.Kind {
	case domain.AppEventLifecycle, domain.AppEventMarker:
		if event.Passed != nil {
			return fmt.Errorf("passed is valid only for assertion events")
		}
	case domain.AppEventAssertion:
		if event.Passed == nil {
			return fmt.Errorf("assertion events require passed")
		}
	default:
		return fmt.Errorf("kind must be lifecycle, marker, or assertion")
	}
	event.Name = strings.TrimSpace(event.Name)
	if !eventNamePattern.MatchString(event.Name) {
		return fmt.Errorf("name must contain 1-64 letters, digits, dots, underscores, or hyphens")
	}
	event.SessionID = strings.TrimSpace(event.SessionID)
	if len(event.SessionID) > 128 {
		return fmt.Errorf("sessionId cannot exceed 128 bytes")
	}
	return nil
}

func sanitizedAttributes(attributes map[string]any) map[string]any {
	if attributes == nil {
		return nil
	}
	result, _ := sandbox.RedactValue(attributes).(map[string]any)
	return result
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("request must contain one JSON object")
	}
	return err
}
