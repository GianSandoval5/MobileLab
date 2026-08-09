package sandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

const maxCapturedBody = 1 << 20

type Handler struct {
	config       config.Config
	fixturesRoot string
	state        *RuntimeState
	requests     domain.RequestRepository
	endpoints    map[string]config.EndpointDefinition
	tokens       *tokenIssuer
	events       domain.EventPublisher
	recorder     domain.CaptureSink
	onError      func(error)
}

func (h *Handler) SetEventPublisher(publisher domain.EventPublisher) {
	h.events = publisher
}

func (h *Handler) SetCaptureSink(recorder domain.CaptureSink) {
	h.recorder = recorder
}

func (h *Handler) SetErrorHandler(handler func(error)) {
	h.onError = handler
}

func NewHandler(cfg config.Config, workspace string, state *RuntimeState, requests domain.RequestRepository) (*Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if state == nil || requests == nil {
		return nil, fmt.Errorf("sandbox state and request repository are required")
	}
	tokens, err := newTokenIssuer()
	if err != nil {
		return nil, err
	}
	endpoints := make(map[string]config.EndpointDefinition, len(cfg.Endpoints))
	for _, endpoint := range cfg.Endpoints {
		endpoints[strings.ToUpper(endpoint.Method)+" "+endpoint.Path] = endpoint
	}
	return &Handler{
		config: cfg, fixturesRoot: filepath.Join(workspace, "fixtures"), state: state,
		requests: requests, endpoints: endpoints, tokens: tokens,
	}, nil
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	started := time.Now()
	statusRecorder := &responseRecorder{ResponseWriter: writer, status: http.StatusOK}
	body := h.captureBody(request)

	switch request.URL.Path {
	case "/__mobilelab/auth/login":
		h.handleLogin(statusRecorder, request)
	case "/__mobilelab/auth/refresh":
		h.handleRefresh(statusRecorder, request)
	default:
		h.handleMock(statusRecorder, request)
	}

	record := domain.RequestRecord{
		Method: request.Method, Path: request.URL.Path,
		Query: RedactQuery(request.URL.Query()), Headers: RedactHeaders(request.Header), Body: body,
		Status: statusRecorder.status, ResponseHeaders: RedactHeaders(statusRecorder.Header()), ResponseBody: statusRecorder.capturedBody(),
		DurationMS: time.Since(started).Milliseconds(), Timestamp: started.UTC(),
	}
	if err := h.requests.Append(request.Context(), record); err != nil {
		h.reportError(fmt.Errorf("record request: %w", err))
		return
	}
	if h.events != nil {
		event := domain.Event{Type: domain.EventRequestRecorded, Version: 1, Timestamp: record.Timestamp, Payload: record}
		if err := h.events.Publish(request.Context(), event); err != nil {
			h.reportError(fmt.Errorf("publish request event: %w", err))
		}
	}
	if h.recorder != nil {
		if err := h.recorder.RecordCapture(request.Context(), domain.CaptureEvent{Kind: domain.CaptureHTTPExchange, Timestamp: record.Timestamp, HTTP: &record}); err != nil {
			h.reportError(fmt.Errorf("capture HTTP exchange: %w", err))
		}
	}
}

func (h *Handler) reportError(err error) {
	if h.onError != nil {
		h.onError(err)
	}
}

func (h *Handler) handleMock(writer http.ResponseWriter, request *http.Request) {
	endpoint, found := h.findEndpoint(request.Method, request.URL.Path)
	if !found {
		writeJSON(writer, http.StatusNotFound, map[string]any{"error": "no mock configured", "method": request.Method, "path": request.URL.Path})
		return
	}
	snapshot := h.state.Snapshot()
	delay := snapshot.LatencyMS + endpoint.DelayMS
	if delay > 0 {
		timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
		defer timer.Stop()
		select {
		case <-request.Context().Done():
			return
		case <-timer.C:
		}
	}
	if snapshot.ForcedError != 0 {
		writeJSON(writer, snapshot.ForcedError, map[string]any{"error": "forced MobileLab response", "status": snapshot.ForcedError})
		return
	}
	if h.config.Sandbox.ErrorRate > 0 && rand.IntN(100) < h.config.Sandbox.ErrorRate {
		writeJSON(writer, http.StatusInternalServerError, map[string]any{"error": "simulated MobileLab error"})
		return
	}
	if endpoint.Error != nil {
		writeJSON(writer, endpoint.Error.Status, map[string]any{"error": "configured endpoint error", "status": endpoint.Error.Status})
		return
	}
	if endpoint.Protected {
		if snapshot.AuthExpired || h.validateBearer(request) != nil {
			writeJSON(writer, http.StatusUnauthorized, map[string]any{"error": "authentication required"})
			return
		}
	}

	for name, value := range endpoint.Response.Headers {
		writer.Header().Set(name, value)
	}
	if endpoint.Response.Fixture != "" {
		body, err := loadFixture(h.fixturesRoot, endpoint.Response.Fixture, h.config.Variables)
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(endpoint.Response.Status)
		_, _ = writer.Write(body)
		return
	}
	writeJSON(writer, endpoint.Response.Status, renderVariables(endpoint.Response.Body, h.config.Variables))
}

func (h *Handler) findEndpoint(method, path string) (config.EndpointDefinition, bool) {
	if endpoint, found := h.endpoints[method+" "+path]; found {
		return endpoint, true
	}
	for key, endpoint := range h.endpoints {
		if !strings.HasPrefix(key, method+" ") {
			continue
		}
		if pathMatches(endpoint.Path, path) {
			return endpoint, true
		}
	}
	return config.EndpointDefinition{}, false
}

func pathMatches(pattern, path string) bool {
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	if len(patternSegments) != len(pathSegments) {
		return false
	}
	for index := range patternSegments {
		segment := patternSegments[index]
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") && len(segment) > 2 {
			if pathSegments[index] == "" {
				return false
			}
			continue
		}
		if segment != pathSegments[index] {
			return false
		}
	}
	return true
}

func (h *Handler) handleLogin(writer http.ResponseWriter, request *http.Request) {
	if !h.config.Auth.Enabled {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "auth sandbox disabled"})
		return
	}
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)
		writeJSON(writer, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
		return
	}
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(request.Body).Decode(&credentials); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid JSON credentials"})
		return
	}
	if credentials.Username != "mobilelab" || credentials.Password != "mobilelab" {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "invalid local credentials"})
		return
	}
	access, err := h.tokens.issue(credentials.Username, "access", time.Hour)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "unable to issue local access token"})
		return
	}
	refresh, err := h.tokens.issue(credentials.Username, "refresh", 24*time.Hour)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "unable to issue local refresh token"})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"access_token": access, "refresh_token": refresh, "token_type": "Bearer", "expires_in": 3600})
}

func (h *Handler) handleRefresh(writer http.ResponseWriter, request *http.Request) {
	if !h.config.Auth.Enabled {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "auth sandbox disabled"})
		return
	}
	var payload struct {
		RefreshToken string `json:"refresh_token"`
	}
	if request.Method != http.MethodPost || json.NewDecoder(request.Body).Decode(&payload) != nil || h.tokens.validate(payload.RefreshToken, "refresh") != nil {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		return
	}
	access, err := h.tokens.issue("mobilelab", "access", time.Hour)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "unable to issue local access token"})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"access_token": access, "token_type": "Bearer", "expires_in": 3600})
}

func (h *Handler) validateBearer(request *http.Request) error {
	value := request.Header.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return fmt.Errorf("bearer token missing")
	}
	return h.tokens.validate(strings.TrimPrefix(value, "Bearer "), "access")
}

func (h *Handler) captureBody(request *http.Request) any {
	if request.Body == nil {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(request.Body, maxCapturedBody))
	if err != nil {
		return "[UNREADABLE]"
	}
	request.Body.Close()
	request.Body = io.NopCloser(strings.NewReader(string(data)))
	if len(data) == 0 {
		return nil
	}
	var decoded any
	if json.Unmarshal(data, &decoded) == nil {
		return RedactValue(decoded)
	}
	return "[NON-JSON BODY]"
}

func RedactQuery(query map[string][]string) map[string][]string {
	return RedactHeaders(query)
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(writer).Encode(value)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status      int
	body        bytes.Buffer
	truncated   bool
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
		r.status = http.StatusOK
	}
	remaining := maxCapturedBody - r.body.Len()
	if remaining > 0 {
		captured := data
		if len(captured) > remaining {
			captured = captured[:remaining]
			r.truncated = true
		}
		_, _ = r.body.Write(captured)
	} else if len(data) > 0 {
		r.truncated = true
	}
	return r.ResponseWriter.Write(data)
}

func (r *responseRecorder) capturedBody() any {
	if r.truncated {
		return "[TRUNCATED RESPONSE BODY]"
	}
	if r.body.Len() == 0 {
		return nil
	}
	var decoded any
	if json.Unmarshal(r.body.Bytes(), &decoded) == nil {
		return RedactValue(decoded)
	}
	return "[NON-JSON BODY]"
}
