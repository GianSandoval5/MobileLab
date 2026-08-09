package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/events"
	"github.com/mobilelab-dev/mobilelab/internal/storage"
)

func TestHandlerServesFixtureAndRecordsRedactedRequest(t *testing.T) {
	handler, state, repository := testHandler(t, false)
	state.SetLatency(20)
	request := httptest.NewRequest(http.MethodGet, "/api/profile?token=secret&view=full", bytes.NewBufferString(`{"password":"secret","safe":"value"}`))
	request.Header.Set("Authorization", "Bearer secret")
	response := httptest.NewRecorder()

	started := time.Now()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"id":"123"`) {
		t.Fatalf("unexpected response %d: %s", response.Code, response.Body.String())
	}
	if time.Since(started) < 20*time.Millisecond {
		t.Fatal("configured latency was not applied")
	}
	records, err := repository.Recent(context.Background(), 1)
	if err != nil || len(records) != 1 {
		t.Fatalf("unexpected records: %v, %v", records, err)
	}
	record := records[0]
	if record.Headers["Authorization"][0] != redacted || record.Query["token"][0] != redacted {
		t.Fatalf("request metadata was not redacted: %#v", record)
	}
	body := record.Body.(map[string]any)
	if body["password"] != redacted || body["safe"] != "value" {
		t.Fatalf("request body was not redacted correctly: %#v", body)
	}
	if record.ResponseHeaders["Content-Type"][0] != "application/json" || record.ResponseBody.(map[string]any)["id"] != "123" {
		t.Fatalf("response was not captured: %#v", record)
	}
}

func TestHandlerAppliesAndResetsForcedError(t *testing.T) {
	handler, state, _ := testHandler(t, false)
	state.SetError(http.StatusServiceUnavailable)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/profile", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", response.Code)
	}

	state.Reset()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/profile", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("got %d after reset, want 200", response.Code)
	}
}

func TestHandlerMatchesOpenAPIPathParameters(t *testing.T) {
	handler, _, _ := testHandler(t, false)
	handler.endpoints["GET /api/users/{id}"] = config.EndpointDefinition{
		Method: "GET", Path: "/api/users/{id}", Response: config.EndpointResponse{Status: http.StatusOK, Body: map[string]any{"found": true}},
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/users/123", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"found":true`) {
		t.Fatalf("parameterized route failed: %d %s", response.Code, response.Body.String())
	}
}

func TestHandlerPublishesSanitizedRequestEvent(t *testing.T) {
	handler, _, _ := testHandler(t, false)
	bus := events.NewBus(2)
	defer bus.Close()
	stream, cancel, err := bus.Subscribe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	handler.SetEventPublisher(bus)
	request := httptest.NewRequest(http.MethodGet, "/api/profile", strings.NewReader(`{"token":"secret"}`))
	request.Header.Set("Authorization", "Bearer secret")
	handler.ServeHTTP(httptest.NewRecorder(), request)
	select {
	case event := <-stream:
		if event.Type != domain.EventRequestRecorded {
			t.Fatalf("unexpected event: %#v", event)
		}
		record := event.Payload.(domain.RequestRecord)
		if record.Headers["Authorization"][0] != redacted || record.Body.(map[string]any)["token"] != redacted {
			t.Fatalf("event was not sanitized: %#v", record)
		}
	case <-time.After(time.Second):
		t.Fatal("request event not published")
	}
}

func TestProtectedEndpointUsesLocalJWTAndExpiryState(t *testing.T) {
	handler, state, _ := testHandler(t, true)

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/profile", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", unauthorized.Code)
	}

	login := httptest.NewRecorder()
	handler.ServeHTTP(login, httptest.NewRequest(http.MethodPost, "/__mobilelab/auth/login", strings.NewReader(`{"username":"mobilelab","password":"mobilelab"}`)))
	if login.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", login.Code, login.Body.String())
	}
	var tokens map[string]any
	if err := json.Unmarshal(login.Body.Bytes(), &tokens); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	request.Header.Set("Authorization", "Bearer "+tokens["access_token"].(string))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("authorized request failed: %d %s", response.Code, response.Body.String())
	}

	state.SetAuthExpired(true)
	request = httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	request.Header.Set("Authorization", "Bearer "+tokens["access_token"].(string))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expired auth got %d, want 401", response.Code)
	}
}

func testHandler(t *testing.T, protected bool) (*Handler, *RuntimeState, *storage.MemoryRequests) {
	t.Helper()
	workspace := t.TempDir()
	fixtures := filepath.Join(workspace, "fixtures")
	if err := os.MkdirAll(fixtures, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtures, "profile.json"), []byte(`{"id":"{{userId}}"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default("test")
	cfg.Endpoints[0].Protected = protected
	state := NewRuntimeState(0)
	repository := storage.NewMemoryRequests(20)
	handler, err := NewHandler(cfg, workspace, state, repository)
	if err != nil {
		t.Fatal(err)
	}
	return handler, state, repository
}

var _ domain.RequestRepository = (*storage.MemoryRequests)(nil)
