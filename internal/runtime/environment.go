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
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	eventbus "github.com/mobilelab-dev/mobilelab/internal/events"
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
	events     *eventbus.Bus
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
		state: sandbox.NewRuntimeState(cfg.Sandbox.LatencyMS), events: eventbus.NewBus(128),
	}, nil
}

func (e *Environment) SetHeadless(headless bool) {
	if headless {
		e.config.Dashboard.Enabled = false
	}
}

func (e *Environment) Run(ctx context.Context) error {
	databasePath := filepath.Join(e.config.Workspace(e.configPath), "mobilelab.db")
	store, err := storage.OpenSQLite(databasePath)
	if err != nil {
		return fmt.Errorf("open MobileLab persistence: %w", err)
	}
	defer store.Close()
	defer e.events.Close()
	e.requests = store
	e.runs = store.ScenarioRuns()
	e.appEvents = store

	handler, err := sandbox.NewHandler(e.config, e.config.Workspace(e.configPath), e.state, e.requests)
	if err != nil {
		return fmt.Errorf("create API sandbox: %w", err)
	}
	handler.SetEventPublisher(e.events)
	handler.SetErrorHandler(func(err error) { fmt.Fprintf(e.out, "! %v\n", err) })
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
			e.publishState(request.Context())
			writer.WriteHeader(http.StatusNoContent)
		case "reset":
			e.state.Reset()
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
		default:
			http.NotFound(writer, request)
		}
	})
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
