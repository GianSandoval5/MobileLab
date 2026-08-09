package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type Status struct {
	PID          int       `json:"pid"`
	StartedAt    time.Time `json:"started_at"`
	Uptime       string    `json:"uptime"`
	LatencyMS    int       `json:"latency_ms"`
	Error        int       `json:"forced_error,omitempty"`
	AuthExpired  bool      `json:"auth_expired"`
	Requests     int       `json:"requests"`
	ScenarioRuns int       `json:"scenario_runs"`
}

type Client struct {
	ConfigPath string
}

func (c Client) SetLatency(ctx context.Context, milliseconds int) error {
	return SetLatency(ctx, c.ConfigPath, milliseconds)
}

func (c Client) SetError(ctx context.Context, status int) error {
	return SetError(ctx, c.ConfigPath, status)
}

func (c Client) SetAuthExpired(ctx context.Context, expired bool) error {
	return SetAuthExpired(ctx, c.ConfigPath, expired)
}

func (c Client) Reset(ctx context.Context) error {
	return ResetFaults(ctx, c.ConfigPath)
}

func (c Client) RecentRequests(ctx context.Context, limit int) ([]domain.RequestRecord, error) {
	var records []domain.RequestRecord
	if err := controlRequest(ctx, c.ConfigPath, http.MethodGet, fmt.Sprintf("requests?limit=%d", limit), nil, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (c Client) Save(ctx context.Context, result domain.ScenarioResult) error {
	return controlRequest(ctx, c.ConfigPath, http.MethodPost, "scenario-runs", result, nil)
}

func (c Client) Recent(ctx context.Context, limit int) ([]domain.ScenarioResult, error) {
	var results []domain.ScenarioResult
	if err := controlRequest(ctx, c.ConfigPath, http.MethodGet, fmt.Sprintf("scenario-runs?limit=%d", limit), nil, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func GetStatus(ctx context.Context, configPath string) (Status, error) {
	var status Status
	if err := controlRequest(ctx, configPath, http.MethodGet, "status", nil, &status); err != nil {
		return Status{}, err
	}
	return status, nil
}

func Stop(ctx context.Context, configPath string) error {
	return controlRequest(ctx, configPath, http.MethodPost, "stop", nil, nil)
}

func SetLatency(ctx context.Context, configPath string, milliseconds int) error {
	return controlRequest(ctx, configPath, http.MethodPost, "latency", map[string]int{"milliseconds": milliseconds}, nil)
}

func SetError(ctx context.Context, configPath string, status int) error {
	return controlRequest(ctx, configPath, http.MethodPost, "error", map[string]int{"status": status}, nil)
}

func ResetFaults(ctx context.Context, configPath string) error {
	return controlRequest(ctx, configPath, http.MethodPost, "reset", nil, nil)
}

func SetAuthExpired(ctx context.Context, configPath string, expired bool) error {
	return controlRequest(ctx, configPath, http.MethodPost, "auth", map[string]bool{"expired": expired}, nil)
}

func controlRequest(ctx context.Context, configPath, method, action string, input, output any) error {
	state, err := ReadState(configPath)
	if err != nil {
		return err
	}
	var body io.Reader
	if input != nil {
		data, marshalErr := json.Marshal(input)
		if marshalErr != nil {
			return marshalErr
		}
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, "http://"+state.Address+"/__mobilelab/control/"+action, body)
	if err != nil {
		return err
	}
	request.Header.Set("X-MobileLab-Control-Token", state.Token)
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 3 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("MobileLab state exists but the process is unreachable at %s: %w", state.Address, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		data, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("control request failed (%s): %s", response.Status, bytes.TrimSpace(data))
	}
	if output != nil {
		if err := json.NewDecoder(response.Body).Decode(output); err != nil {
			return fmt.Errorf("decode control response: %w", err)
		}
	}
	return nil
}
