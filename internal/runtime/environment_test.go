package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/datastore"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/storage"
)

func TestEnvironmentLifecycleAndRuntimeFaults(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, config.DefaultFilename)
	cfg := config.Default("integration")
	cfg.Server.Port = availablePort(t)
	cfg.Dashboard.Enabled = true
	if err := os.MkdirAll(filepath.Join(root, "mobilelab", "fixtures"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "mobilelab", "fixtures", "profile.json"), []byte(`{"id":"123"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "mobilelab", "seeds"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "mobilelab", "seeds", "items.json"), []byte(`[{"id":"seeded","name":"Seeded"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "mobilelab", "data.yaml"), []byte("schema_version: 1\nresources:\n  items:\n    path: /api/items\n    id: id\n    seed: seeds/items.json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := config.Write(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	environment, err := NewEnvironment(configPath, &output)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- environment.Run(ctx) }()
	waitForState(t, configPath)
	dashboardResponse := get(t, "http://"+cfg.Address()+"/dashboard")
	if dashboardResponse.StatusCode != http.StatusOK {
		t.Fatalf("dashboard returned %d", dashboardResponse.StatusCode)
	}
	dashboardResponse.Body.Close()
	connection, _, err := websocket.Dial(context.Background(), "ws://"+cfg.Address()+"/__mobilelab/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventSnapshot || !strings.Contains(fmt.Sprint(event.Payload), "items") {
		t.Fatalf("dashboard did not send initial snapshot: %#v", event)
	}
	sdkBody := strings.NewReader(`{"protocolVersion":1,"framework":"flutter","kind":"lifecycle","name":"ready","attributes":{"token":"secret"}}`)
	sdkResponse, err := http.Post("http://"+cfg.Address()+"/__mobilelab/sdk/events", "application/json", sdkBody)
	if err != nil {
		t.Fatal(err)
	}
	if sdkResponse.StatusCode != http.StatusAccepted {
		data, _ := io.ReadAll(sdkResponse.Body)
		t.Fatalf("SDK bridge returned %d: %s", sdkResponse.StatusCode, data)
	}
	sdkResponse.Body.Close()
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventAppReported {
		t.Fatalf("dashboard did not receive app event: %#v", event)
	}
	appEvents, err := (Client{ConfigPath: configPath}).RecentAppEvents(context.Background(), 10)
	if err != nil || len(appEvents) != 1 || appEvents[0].Name != "ready" || appEvents[0].Attributes["token"] != "[REDACTED]" {
		t.Fatalf("runtime app event observation failed: %#v, %v", appEvents, err)
	}

	status, err := GetStatus(context.Background(), configPath)
	if err != nil || status.PID < 1 {
		t.Fatalf("status failed: %#v, %v", status, err)
	}
	response := get(t, "http://"+cfg.Address()+"/api/profile")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("mock returned %d", response.StatusCode)
	}
	response.Body.Close()
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventRequestRecorded {
		t.Fatalf("dashboard did not receive request event: %#v", event)
	}
	created, err := http.Post("http://"+cfg.Address()+"/api/items", "application/json", strings.NewReader(`{"id":"persisted","name":"Created"}`))
	if err != nil {
		t.Fatal(err)
	}
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("database create returned %d", created.StatusCode)
	}
	created.Body.Close()
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventRequestRecorded {
		t.Fatalf("database create did not publish request: %#v", event)
	}
	databaseResponse := get(t, "http://"+cfg.Address()+"/api/items/persisted")
	if databaseResponse.StatusCode != http.StatusOK {
		t.Fatalf("database read returned %d", databaseResponse.StatusCode)
	}
	databaseResponse.Body.Close()
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventRequestRecorded {
		t.Fatalf("database read did not publish request: %#v", event)
	}
	recent, err := (Client{ConfigPath: configPath}).RecentRequests(context.Background(), 10)
	if err != nil || len(recent) != 3 || recent[0].Path != "/api/profile" || recent[1].Path != "/api/items" || recent[2].Path != "/api/items/persisted" {
		t.Fatalf("runtime request observation failed: %#v, %v", recent, err)
	}
	client := Client{ConfigPath: configPath}
	startedRecording, err := client.StartRecording(context.Background(), "login")
	if err != nil || startedRecording.Name != "login" {
		t.Fatalf("start recording: %#v, %v", startedRecording, err)
	}
	if err := SetLatency(context.Background(), configPath, 125); err != nil {
		t.Fatal(err)
	}
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventStateChanged {
		t.Fatalf("recorded latency did not publish state: %#v", event)
	}
	if err := client.RecordCapture(context.Background(), domain.CaptureEvent{Kind: domain.CaptureDeepLink, DeepLink: &domain.DeepLinkCapture{URL: "myapp://login"}}); err != nil {
		t.Fatal(err)
	}
	response = get(t, "http://"+cfg.Address()+"/api/profile")
	response.Body.Close()
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventRequestRecorded {
		t.Fatalf("recorded HTTP did not publish request: %#v", event)
	}
	captured, err := client.StopRecording(context.Background())
	if err != nil || len(captured.Events) != 3 {
		t.Fatalf("stop recording: %#v, %v", captured, err)
	}
	if captured.Events[0].Kind != domain.CaptureEnvironment || captured.Events[1].Kind != domain.CaptureDeepLink || captured.Events[2].Kind != domain.CaptureHTTPExchange {
		t.Fatalf("recording order changed: %#v", captured.Events)
	}
	if captured.Events[2].HTTP.ResponseBody.(map[string]any)["id"] != "123" {
		t.Fatalf("response body missing from recording: %#v", captured.Events[2])
	}
	run := domain.ScenarioResult{Name: "Persistent run", Passed: true, StartedAt: time.Now().UTC(), DurationMS: 12}
	if err := client.Save(context.Background(), run); err != nil {
		t.Fatalf("save scenario run: %v", err)
	}
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventScenarioCompleted {
		t.Fatalf("dashboard did not receive scenario event: %#v", event)
	}
	runs, err := client.Recent(context.Background(), 10)
	if err != nil || len(runs) != 1 || runs[0].Name != run.Name {
		t.Fatalf("runtime scenario history failed: %#v, %v", runs, err)
	}
	status, err = GetStatus(context.Background(), configPath)
	if err != nil || status.ScenarioRuns != 1 || status.AppEvents != 1 {
		t.Fatalf("status did not include scenario history: %#v, %v", status, err)
	}

	if err := SetError(context.Background(), configPath, http.StatusBadGateway); err != nil {
		t.Fatal(err)
	}
	if event := readRuntimeEvent(t, connection); event.Type != domain.EventStateChanged {
		t.Fatalf("dashboard did not receive state event: %#v", event)
	}
	response = get(t, "http://"+cfg.Address()+"/api/profile")
	if response.StatusCode != http.StatusBadGateway {
		t.Fatalf("forced error returned %d", response.StatusCode)
	}
	response.Body.Close()

	if err := ResetFaults(context.Background(), configPath); err != nil {
		t.Fatal(err)
	}
	if err := Stop(context.Background(), configPath); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("environment failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("environment did not stop gracefully")
	}
	if _, err := os.Stat(StatePath(configPath)); !os.IsNotExist(err) {
		t.Fatalf("runtime state was not removed: %v", err)
	}
	store, err := storage.OpenSQLite(filepath.Join(root, "mobilelab", "mobilelab.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	persistedRequests, err := store.Recent(context.Background(), 10)
	if err != nil || len(persistedRequests) < 2 {
		t.Fatalf("request history did not survive restart: %#v, %v", persistedRequests, err)
	}
	persistedRuns, err := store.ScenarioRuns().Recent(context.Background(), 10)
	if err != nil || len(persistedRuns) != 1 || persistedRuns[0].Name != run.Name {
		t.Fatalf("scenario history did not survive restart: %#v, %v", persistedRuns, err)
	}
	persistedAppEvents, err := store.RecentAppEvents(context.Background(), 10)
	if err != nil || len(persistedAppEvents) != 1 || persistedAppEvents[0].Name != "ready" {
		t.Fatalf("app event history did not survive restart: %#v, %v", persistedAppEvents, err)
	}
	dataStore, err := datastore.Open(datastore.DatabasePath(root))
	if err != nil {
		t.Fatal(err)
	}
	defer dataStore.Close()
	document, found, err := dataStore.Get(context.Background(), "items", "persisted")
	if err != nil || !found || document["name"] != "Created" {
		t.Fatalf("business document did not survive shutdown: %#v, %v", document, err)
	}
}

func TestLoopbackOnlyRejectsRemoteDashboardClients(t *testing.T) {
	called := false
	handler := loopbackOnly(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		called = true
		writer.WriteHeader(http.StatusNoContent)
	}))
	remote := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	remote.RemoteAddr = "192.0.2.10:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, remote)
	if response.Code != http.StatusForbidden || called {
		t.Fatalf("remote dashboard client was accepted: code=%d called=%v", response.Code, called)
	}

	local := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	local.RemoteAddr = "127.0.0.1:1234"
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, local)
	if response.Code != http.StatusNoContent || !called {
		t.Fatalf("loopback dashboard client was rejected: code=%d called=%v", response.Code, called)
	}
}

func readRuntimeEvent(t *testing.T, connection *websocket.Conn) domain.Event {
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

func availablePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitForState(t *testing.T, configPath string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := ReadState(configPath); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("runtime state was not created")
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, response.Body)
	return response
}
