package runtime

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/dashboard"
	"github.com/mobilelab-dev/mobilelab/internal/datastore"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	eventbus "github.com/mobilelab-dev/mobilelab/internal/events"
	"github.com/mobilelab-dev/mobilelab/internal/recording"
	"github.com/mobilelab-dev/mobilelab/internal/sandbox"
	"github.com/mobilelab-dev/mobilelab/internal/sdkbridge"
	"github.com/mobilelab-dev/mobilelab/internal/storage"
)

type Environment struct {
	configPath string
	config     config.Config
	out        io.Writer
	state      *sandbox.RuntimeState
	requests   domain.RequestRepository
	runs       domain.ScenarioRunRepository
	appEvents  domain.AppEventRepository
	data       *datastore.Store
	dataConfig *datastore.Config
	events     *eventbus.Bus
	recorder   *recording.Service
	startedAt  time.Time
	stop       context.CancelFunc
}

func NewEnvironment(configPath string, out io.Writer) (*Environment, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	return &Environment{
		configPath: configPath, config: cfg, out: out,
		state: sandbox.NewRuntimeState(cfg.Sandbox.LatencyMS), events: eventbus.NewBus(128), recorder: recording.NewService(),
	}, nil
}

func (e *Environment) SetHeadless(headless bool) {
	if headless {
		e.config.Dashboard.Enabled = false
	}
}

func (e *Environment) Run(ctx context.Context) error {
	workspace := e.config.Workspace(e.configPath)
	databasePath := filepath.Join(workspace, "mobilelab.db")
	store, err := storage.OpenSQLite(databasePath)
	if err != nil {
		return fmt.Errorf("open MobileLab persistence: %w", err)
	}
	defer store.Close()
	defer e.events.Close()
	e.requests = store
	e.runs = store.ScenarioRuns()
	e.appEvents = store

	dataConfig, configured, err := datastore.LoadOptional(filepath.Dir(e.configPath))
	if err != nil {
		return fmt.Errorf("load business data API: %w", err)
	}
	if configured {
		if err := dataConfig.ValidateEndpoints(e.config.Endpoints); err != nil {
			return fmt.Errorf("validate business data API: %w", err)
		}
		dataStore, openErr := datastore.Open(datastore.DatabasePath(filepath.Dir(e.configPath)))
		if openErr != nil {
			return fmt.Errorf("open business data API: %w", openErr)
		}
		defer dataStore.Close()
		if seedErr := dataStore.Seed(ctx, dataConfig, workspace, datastore.SeedEmpty); seedErr != nil {
			return fmt.Errorf("initialize business data API: %w", seedErr)
		}
		e.data = dataStore
		e.dataConfig = &dataConfig
	}

	handler, err := sandbox.NewHandler(e.config, workspace, e.state, e.requests)
	if err != nil {
		return fmt.Errorf("create API sandbox: %w", err)
	}
	handler.SetEventPublisher(e.events)
	handler.SetCaptureSink(e.recorder)
	handler.SetErrorHandler(func(err error) { fmt.Fprintf(e.out, "! %v\n", err) })
	if configured {
		dataHandler, dataErr := datastore.NewHandler(dataConfig, e.data)
		if dataErr != nil {
			return fmt.Errorf("create business data API: %w", dataErr)
		}
		handler.SetDynamicHandler(dataHandler)
	}
	listener, err := net.Listen("tcp", e.config.Address())
	if err != nil {
		return fmt.Errorf("unable to start MobileLab: address %s is unavailable: %w", e.config.Address(), err)
	}
	defer listener.Close()

	runContext, cancel := context.WithCancel(ctx)
	e.stop = cancel
	defer cancel()
	e.startedAt = time.Now().UTC()
	token, err := randomToken()
	if err != nil {
		return err
	}
	instance := InstanceState{PID: os.Getpid(), Address: listener.Addr().String(), Token: token, StartedAt: e.startedAt}
	if err := writeState(e.configPath, instance); err != nil {
		return err
	}
	defer func() { _ = removeOwnedState(e.configPath, token) }()

	mux := http.NewServeMux()
	mux.Handle("/__mobilelab/control/", e.controlHandler(token))
	mux.Handle("/__mobilelab/sdk/events", sdkbridge.Handler{Repository: e.appEvents, Events: e.events})
	if e.config.Dashboard.Enabled {
		dashboardServer := dashboard.Server{Bus: e.events, Requests: e.requests, Runs: e.runs, AppEvents: e.appEvents, State: func() any { return e.state.Snapshot() }}
		if e.data != nil && e.dataConfig != nil {
			dashboardServer.Data = func(ctx context.Context) any {
				counts, countErr := e.data.Counts(ctx, *e.dataConfig)
				if countErr != nil {
					return map[string]any{"error": "unavailable"}
				}
				return map[string]any{"resources": counts}
			}
		}
		mux.Handle("/dashboard", loopbackOnly(http.HandlerFunc(dashboardServer.Page)))
		mux.Handle("/__mobilelab/events", loopbackOnly(http.HandlerFunc(dashboardServer.Events)))
	}
	mux.Handle("/", handler)
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	fmt.Fprintf(e.out, "MobileLab environment ready\n✓ API Sandbox  http://%s\n", listener.Addr())
	if configured {
		counts, countErr := e.data.Counts(ctx, dataConfig)
		if countErr != nil {
			return fmt.Errorf("inspect business data API: %w", countErr)
		}
		fmt.Fprintf(e.out, "✓ Data API     %d resources · %d documents · mobilelab/%s\n", len(counts), totalDocuments(counts), datastore.DatabaseFilename)
	}
	if e.config.Server.Host != "localhost" {
		if ip := net.ParseIP(e.config.Server.Host); ip != nil && !ip.IsLoopback() {
			fmt.Fprintln(e.out, "! Warning: MobileLab is exposed beyond loopback. Use only on a trusted development network.")
		}
	}
	if e.config.Dashboard.Enabled {
		fmt.Fprintf(e.out, "✓ Dashboard    http://%s/dashboard\n", listener.Addr())
	}
	fmt.Fprintln(e.out, "Press Ctrl+C to stop.")

	select {
	case <-runContext.Done():
	case serveErr := <-serverErrors:
		if !errors.Is(serveErr, http.ErrServerClosed) {
			return fmt.Errorf("MobileLab server failed: %w", serveErr)
		}
	}
	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shut down MobileLab: %w", err)
	}
	fmt.Fprintln(e.out, "MobileLab stopped.")
	return nil
}

func totalDocuments(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func (e *Environment) controlHandler(token string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-MobileLab-Control-Token") != token {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		action := strings.TrimPrefix(request.URL.Path, "/__mobilelab/control/")
		if action == "status" && request.Method == http.MethodGet {
			recent, err := e.requests.Recent(request.Context(), 500)
			if err != nil {
				http.Error(writer, "unable to read request history", http.StatusInternalServerError)
				return
			}
			runs, err := e.runs.Recent(request.Context(), 500)
			if err != nil {
				http.Error(writer, "unable to read scenario history", http.StatusInternalServerError)
				return
			}
			appEvents, err := e.appEvents.RecentAppEvents(request.Context(), 500)
			if err != nil {
				http.Error(writer, "unable to read app event history", http.StatusInternalServerError)
				return
			}
			snapshot := e.state.Snapshot()
			writeControlJSON(writer, Status{
				PID: os.Getpid(), StartedAt: e.startedAt, Uptime: time.Since(e.startedAt).Round(time.Second).String(),
				LatencyMS: snapshot.LatencyMS, Error: snapshot.ForcedError, AuthExpired: snapshot.AuthExpired, Requests: len(recent), ScenarioRuns: len(runs), AppEvents: len(appEvents),
			})
			return
		}
		if action == "requests" && request.Method == http.MethodGet {
			limit := 100
			if value := request.URL.Query().Get("limit"); value != "" {
				if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 && parsed <= 500 {
					limit = parsed
				}
			}
			recent, err := e.requests.Recent(request.Context(), limit)
			if err != nil {
				http.Error(writer, "unable to read request history", http.StatusInternalServerError)
				return
			}
			writeControlJSON(writer, recent)
			return
		}
		if action == "scenario-runs" && request.Method == http.MethodGet {
			limit := queryLimit(request, 100)
			recent, err := e.runs.Recent(request.Context(), limit)
			if err != nil {
				http.Error(writer, "unable to read scenario history", http.StatusInternalServerError)
				return
			}
			writeControlJSON(writer, recent)
			return
		}
		if action == "app-events" && request.Method == http.MethodGet {
			recent, err := e.appEvents.RecentAppEvents(request.Context(), queryLimit(request, 100))
			if err != nil {
				http.Error(writer, "unable to read app event history", http.StatusInternalServerError)
				return
			}
			writeControlJSON(writer, recent)
			return
		}
		if action == "recording" && request.Method == http.MethodGet {
			recording, active := e.recorder.ActiveRecording(request.Context())
			writeControlJSON(writer, map[string]any{"active": active, "recording": recording})
			return
		}
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch action {
		case "stop":
			writer.WriteHeader(http.StatusAccepted)
			go e.stop()
		case "latency":
			var input struct {
				Milliseconds int `json:"milliseconds"`
			}
			if json.NewDecoder(request.Body).Decode(&input) != nil || input.Milliseconds < 0 || input.Milliseconds > 300_000 {
				http.Error(writer, "milliseconds must be between 0 and 300000", http.StatusBadRequest)
				return
			}
			e.state.SetLatency(input.Milliseconds)
			e.captureEnvironment(request.Context(), "latency")
			e.publishState(request.Context())
			writer.WriteHeader(http.StatusNoContent)
		case "error":
			var input struct {
				Status int `json:"status"`
			}
			if json.NewDecoder(request.Body).Decode(&input) != nil || input.Status < 400 || input.Status > 599 {
				http.Error(writer, "status must be between 400 and 599", http.StatusBadRequest)
				return
			}
			e.state.SetError(input.Status)
			e.captureEnvironment(request.Context(), "error")
			e.publishState(request.Context())
			writer.WriteHeader(http.StatusNoContent)
		case "auth":
			var input struct {
				Expired bool `json:"expired"`
			}
			if json.NewDecoder(request.Body).Decode(&input) != nil {
				http.Error(writer, "invalid auth state", http.StatusBadRequest)
				return
			}
			e.state.SetAuthExpired(input.Expired)
			e.captureEnvironment(request.Context(), "auth")
			e.publishState(request.Context())
			writer.WriteHeader(http.StatusNoContent)
		case "reset":
			e.state.Reset()
			e.captureEnvironment(request.Context(), "reset")
			e.publishState(request.Context())
			writer.WriteHeader(http.StatusNoContent)
		case "scenario-runs":
			var result domain.ScenarioResult
			decoder := json.NewDecoder(io.LimitReader(request.Body, 1<<20))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&result); err != nil || result.Name == "" {
				http.Error(writer, "invalid scenario result", http.StatusBadRequest)
				return
			}
			if err := e.runs.Save(request.Context(), result); err != nil {
				http.Error(writer, "unable to store scenario result", http.StatusInternalServerError)
				return
			}
			e.publish(request.Context(), domain.Event{Type: domain.EventScenarioCompleted, Version: 1, Timestamp: time.Now().UTC(), Payload: result})
			writer.WriteHeader(http.StatusNoContent)
		case "recording/start":
			var input struct {
				Name string `json:"name"`
			}
			decoder := json.NewDecoder(io.LimitReader(request.Body, 4096))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&input); err != nil {
				http.Error(writer, "invalid recording request", http.StatusBadRequest)
				return
			}
			recording, err := e.recorder.StartRecording(request.Context(), input.Name, e.environmentState())
			if err != nil {
				http.Error(writer, err.Error(), http.StatusConflict)
				return
			}
			writeControlJSON(writer, recording)
		case "recording/stop":
			recording, err := e.recorder.StopRecording(request.Context())
			if err != nil {
				http.Error(writer, err.Error(), http.StatusConflict)
				return
			}
			writeControlJSON(writer, recording)
		case "recording/capture":
			var event domain.CaptureEvent
			decoder := json.NewDecoder(io.LimitReader(request.Body, 64<<10))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&event); err != nil || event.Kind != domain.CaptureDeepLink {
				http.Error(writer, "only a valid deep-link capture is accepted", http.StatusBadRequest)
				return
			}
			if err := e.recorder.RecordCapture(request.Context(), event); err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
			writer.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(writer, request)
		}
	})
}

func (e *Environment) environmentState() domain.EnvironmentState {
	snapshot := e.state.Snapshot()
	return domain.EnvironmentState{LatencyMS: snapshot.LatencyMS, ForcedError: snapshot.ForcedError, AuthExpired: snapshot.AuthExpired}
}

func (e *Environment) captureEnvironment(ctx context.Context, action string) {
	state := e.environmentState()
	mutation := domain.EnvironmentMutation{Action: action, LatencyMS: state.LatencyMS, ForcedError: state.ForcedError, AuthExpired: state.AuthExpired}
	if err := e.recorder.RecordCapture(ctx, domain.CaptureEvent{Kind: domain.CaptureEnvironment, Environment: &mutation}); err != nil {
		fmt.Fprintf(e.out, "! capture environment: %v\n", err)
	}
}

func queryLimit(request *http.Request, defaultValue int) int {
	if value := request.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 && parsed <= 500 {
			return parsed
		}
	}
	return defaultValue
}

func (e *Environment) publishState(ctx context.Context) {
	e.publish(ctx, domain.Event{Type: domain.EventStateChanged, Version: 1, Timestamp: time.Now().UTC(), Payload: e.state.Snapshot()})
}

func (e *Environment) publish(ctx context.Context, event domain.Event) {
	if err := e.events.Publish(ctx, event); err != nil {
		fmt.Fprintf(e.out, "! publish event: %v\n", err)
	}
}

func loopbackOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		host, _, err := net.SplitHostPort(request.RemoteAddr)
		ip := net.ParseIP(host)
		if err != nil || ip == nil || !ip.IsLoopback() {
			http.Error(writer, "dashboard is available only from loopback", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate control token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func writeControlJSON(writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(value)
}
