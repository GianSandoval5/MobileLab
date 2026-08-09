package sdkbridge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/events"
)

type memoryEvents struct{ saved []domain.AppEvent }

func (m *memoryEvents) SaveAppEvent(_ context.Context, event domain.AppEvent) error {
	m.saved = append(m.saved, event)
	return nil
}

func (m *memoryEvents) RecentAppEvents(context.Context, int) ([]domain.AppEvent, error) {
	return append([]domain.AppEvent(nil), m.saved...), nil
}

func TestHandlerAcceptsSanitizesPersistsAndPublishesEvent(t *testing.T) {
	repository := &memoryEvents{}
	bus := events.NewBus(2)
	defer bus.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	subscription, unsubscribe, err := bus.Subscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer unsubscribe()
	now := time.Date(2026, 8, 9, 20, 0, 0, 0, time.UTC)
	handler := Handler{Repository: repository, Events: bus, Now: func() time.Time { return now }}
	body := `{"protocolVersion":1,"framework":"flutter","kind":"assertion","name":"checkout.loaded","passed":true,"sessionId":"run-1","attributes":{"token":"secret","screen":"checkout"}}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/__mobilelab/sdk/events", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
	if len(repository.saved) != 1 || repository.saved[0].Timestamp != now {
		t.Fatalf("unexpected saved events: %#v", repository.saved)
	}
	if repository.saved[0].Attributes["token"] != "[REDACTED]" {
		t.Fatalf("secret was not redacted: %#v", repository.saved[0].Attributes)
	}
	select {
	case published := <-subscription:
		if published.Type != domain.EventAppReported {
			t.Fatalf("unexpected event: %#v", published)
		}
	case <-time.After(time.Second):
		t.Fatal("app event was not published")
	}
}

func TestHandlerAcceptsEverySupportedFramework(t *testing.T) {
	frameworks := []domain.AppFramework{
		domain.FrameworkFlutter,
		domain.FrameworkReactNative,
		domain.FrameworkAndroid,
		domain.FrameworkIOS,
		domain.FrameworkCapacitor,
	}
	for _, framework := range frameworks {
		t.Run(string(framework), func(t *testing.T) {
			repository := &memoryEvents{}
			body := `{"protocolVersion":1,"framework":"` + string(framework) + `","kind":"marker","name":"ready"}`
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/__mobilelab/sdk/events", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			(Handler{Repository: repository}).ServeHTTP(response, request)
			if response.Code != http.StatusAccepted || len(repository.saved) != 1 {
				t.Fatalf("status=%d events=%d body=%q", response.Code, len(repository.saved), response.Body.String())
			}
		})
	}
}

func TestHandlerRejectsInvalidContract(t *testing.T) {
	tests := []string{
		`{"protocolVersion":2,"framework":"flutter","kind":"marker","name":"ready"}`,
		`{"protocolVersion":1,"framework":"unknown","kind":"marker","name":"ready"}`,
		`{"protocolVersion":1,"framework":"flutter","kind":"assertion","name":"ready"}`,
		`{"protocolVersion":1,"framework":"flutter","kind":"marker","name":"bad name"}`,
		`{"protocolVersion":1,"framework":"flutter","kind":"marker","name":"ready","unknown":true}`,
	}
	for _, body := range tests {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		(Handler{Repository: &memoryEvents{}}).ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("body=%s status=%d response=%q", body, response.Code, response.Body.String())
		}
	}
}

func TestHandlerRequiresJSONAndLimitsBody(t *testing.T) {
	for _, test := range []struct {
		body        string
		contentType string
		status      int
	}{
		{body: `{}`, contentType: "text/plain", status: http.StatusUnsupportedMediaType},
		{body: strings.Repeat(" ", MaxEventBytes+1), contentType: "application/json", status: http.StatusBadRequest},
	} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
		request.Header.Set("Content-Type", test.contentType)
		(Handler{Repository: &memoryEvents{}}).ServeHTTP(response, request)
		if response.Code != test.status {
			t.Errorf("contentType=%s status=%d, want %d", test.contentType, response.Code, test.status)
		}
	}
}

func TestHandlerRejectsNonPost(t *testing.T) {
	response := httptest.NewRecorder()
	(Handler{Repository: &memoryEvents{}}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("status=%d headers=%v", response.Code, response.Header())
	}
}
