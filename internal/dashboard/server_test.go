package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/events"
	"github.com/mobilelab-dev/mobilelab/internal/storage"
)

func TestPageIncludesLiveDashboardWithoutInterpolatedData(t *testing.T) {
	response := httptest.NewRecorder()
	(Server{}).Page(response, httptest.NewRequest(http.MethodGet, "/dashboard", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Recent requests") || !strings.Contains(response.Body.String(), "__mobilelab/events") {
		t.Fatalf("unexpected dashboard: %d %s", response.Code, response.Body.String())
	}
}

func TestEventsSendsSnapshotAndLiveEvent(t *testing.T) {
	bus := events.NewBus(4)
	defer bus.Close()
	requests := storage.NewMemoryRequests(10)
	now := time.Now().UTC()
	if err := requests.Append(context.Background(), domain.RequestRecord{Method: "GET", Path: "/profile", Status: 200, Timestamp: now}); err != nil {
		t.Fatal(err)
	}
	runs := &memoryRuns{results: []domain.ScenarioResult{{Name: "Existing", Passed: true, StartedAt: now}}}
	dashboard := Server{Bus: bus, Requests: requests, Runs: runs, State: func() any { return map[string]any{"latency_ms": 10} }}
	server := httptest.NewServer(http.HandlerFunc(dashboard.Events))
	defer server.Close()

	connection, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	initial := readEvent(t, connection)
	if initial.Type != domain.EventSnapshot {
		t.Fatalf("unexpected initial event: %#v", initial)
	}
	live := domain.Event{Type: domain.EventStateChanged, Version: 1, Timestamp: now, Payload: map[string]any{"latency_ms": 20}}
	if err := bus.Publish(context.Background(), live); err != nil {
		t.Fatal(err)
	}
	received := readEvent(t, connection)
	if received.Type != domain.EventStateChanged {
		t.Fatalf("unexpected live event: %#v", received)
	}
}

func readEvent(t *testing.T, connection *websocket.Conn) domain.Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, data, err := connection.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var event domain.Event
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatal(err)
	}
	return event
}

type memoryRuns struct{ results []domain.ScenarioResult }

func (m *memoryRuns) Save(_ context.Context, result domain.ScenarioResult) error {
	m.results = append(m.results, result)
	return nil
}

func (m *memoryRuns) Recent(context.Context, int) ([]domain.ScenarioResult, error) {
	return append([]domain.ScenarioResult(nil), m.results...), nil
}
