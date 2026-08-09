package config

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultFilename = "mobilelab.yaml"

type Config struct {
	Project   ProjectConfig          `yaml:"project"`
	Server    ServerConfig           `yaml:"server"`
	Dashboard DashboardConfig        `yaml:"dashboard"`
	Sandbox   SandboxConfig          `yaml:"sandbox"`
	Auth      AuthConfig             `yaml:"auth"`
	Device    DeviceConfig           `yaml:"device"`
	Variables map[string]string      `yaml:"variables,omitempty"`
	Push      map[string]PushFixture `yaml:"push,omitempty"`
	Endpoints []EndpointDefinition   `yaml:"endpoints"`
}

type ProjectConfig struct {
	Name string `yaml:"name"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DashboardConfig struct {
	Enabled bool `yaml:"enabled"`
}

type SandboxConfig struct {
	LatencyMS int `yaml:"latency"`
	ErrorRate int `yaml:"error_rate"`
}

type AuthConfig struct {
	Enabled bool `yaml:"enabled"`
}

type DeviceConfig struct {
	AutoDetect bool `yaml:"auto_detect"`
}

type PushFixture struct {
	Title string         `yaml:"title"`
	Body  string         `yaml:"body"`
	Data  map[string]any `yaml:"data,omitempty"`
}

type EndpointDefinition struct {
	Path      string           `yaml:"path"`
	Method    string           `yaml:"method"`
	Protected bool             `yaml:"protected,omitempty"`
	DelayMS   int              `yaml:"delay,omitempty"`
	Error     *EndpointError   `yaml:"error,omitempty"`
	Response  EndpointResponse `yaml:"response"`
}

type EndpointError struct {
	Status int `yaml:"status"`
}

type EndpointResponse struct {
	Status  int               `yaml:"status"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    any               `yaml:"body,omitempty"`
	Fixture string            `yaml:"fixture,omitempty"`
}

func Default(projectName string) Config {
	if strings.TrimSpace(projectName) == "" {
		projectName = "mobile-app"
	}
	return Config{
		Project:   ProjectConfig{Name: projectName},
		Server:    ServerConfig{Host: "127.0.0.1", Port: 4566},
		Dashboard: DashboardConfig{Enabled: true},
		Auth:      AuthConfig{Enabled: true},
		Device:    DeviceConfig{AutoDetect: true},
		Variables: map[string]string{"userId": "123"},
		Push: map[string]PushFixture{
			"payment-success": {
				Title: "Payment completed",
				Body:  "Your payment was processed",
				Data:  map[string]any{"transactionId": "ABC123"},
			},
		},
		Endpoints: []EndpointDefinition{{
			Path:   "/api/profile",
			Method: http.MethodGet,
			Response: EndpointResponse{
				Status:  http.StatusOK,
				Fixture: "profile.json",
			},
		}},
	}
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open configuration %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse configuration %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration %q: %w", path, err)
	}
	return cfg, nil
}

func Write(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode configuration: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write configuration %q: %w", path, err)
	}
	return nil
}

func (c Config) Validate() error {
	var problems []error
	if strings.TrimSpace(c.Project.Name) == "" {
		problems = append(problems, errors.New("project.name is required"))
	}
	if net.ParseIP(c.Server.Host) == nil && c.Server.Host != "localhost" {
		problems = append(problems, fmt.Errorf("server.host %q must be an IP address or localhost", c.Server.Host))
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		problems = append(problems, fmt.Errorf("server.port must be between 1 and 65535 (got %d)", c.Server.Port))
	}
	if c.Sandbox.LatencyMS < 0 {
		problems = append(problems, errors.New("sandbox.latency cannot be negative"))
	}
	if c.Sandbox.ErrorRate < 0 || c.Sandbox.ErrorRate > 100 {
		problems = append(problems, errors.New("sandbox.error_rate must be between 0 and 100"))
	}
	for name, fixture := range c.Push {
		if strings.TrimSpace(name) == "" {
			problems = append(problems, errors.New("push fixture name cannot be empty"))
		}
		if strings.TrimSpace(fixture.Title) == "" && strings.TrimSpace(fixture.Body) == "" {
			problems = append(problems, fmt.Errorf("push.%s requires title or body", name))
		}
		if _, reserved := fixture.Data["aps"]; reserved {
			problems = append(problems, fmt.Errorf("push.%s.data key %q is reserved", name, "aps"))
		}
	}

	seen := make(map[string]struct{}, len(c.Endpoints))
	for index, endpoint := range c.Endpoints {
		prefix := fmt.Sprintf("endpoints[%d]", index)
		endpoint.Method = strings.ToUpper(endpoint.Method)
		if !strings.HasPrefix(endpoint.Path, "/") {
			problems = append(problems, fmt.Errorf("%s.path must begin with /", prefix))
		}
		if !validMethod(endpoint.Method) {
			problems = append(problems, fmt.Errorf("%s.method %q is not supported", prefix, endpoint.Method))
		}
		if endpoint.DelayMS < 0 {
			problems = append(problems, fmt.Errorf("%s.delay cannot be negative", prefix))
		}
		if endpoint.Response.Status < 100 || endpoint.Response.Status > 599 {
			problems = append(problems, fmt.Errorf("%s.response.status must be between 100 and 599", prefix))
		}
		if endpoint.Response.Fixture != "" && endpoint.Response.Body != nil {
			problems = append(problems, fmt.Errorf("%s.response cannot define both body and fixture", prefix))
		}
		if endpoint.Error != nil && (endpoint.Error.Status < 400 || endpoint.Error.Status > 599) {
			problems = append(problems, fmt.Errorf("%s.error.status must be between 400 and 599", prefix))
		}
		key := endpoint.Method + " " + routeSignature(endpoint.Path)
		if _, exists := seen[key]; exists {
			problems = append(problems, fmt.Errorf("duplicate endpoint %s", key))
		}
		seen[key] = struct{}{}
	}
	return errors.Join(problems...)
}

func routeSignature(path string) string {
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			segments[index] = "{}"
		}
	}
	return strings.Join(segments, "/")
}

func (c Config) Address() string {
	return net.JoinHostPort(c.Server.Host, fmt.Sprintf("%d", c.Server.Port))
}

func (c Config) Workspace(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "mobilelab")
}

func validMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}
