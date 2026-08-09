package recording

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/schemas"
	"gopkg.in/yaml.v3"
)

func GenerateScenario(recording domain.Recording) (domain.ScenarioDefinition, error) {
	if recording.Name == "" || recording.StartedAt.IsZero() {
		return domain.ScenarioDefinition{}, fmt.Errorf("recording name and start time are required")
	}
	definition := domain.ScenarioDefinition{
		Name:    "Recorded " + recording.Name,
		Backend: domain.ScenarioBackend{LatencyMS: recording.InitialEnvironment.LatencyMS, Error: recording.InitialEnvironment.ForcedError},
	}
	if recording.InitialEnvironment.AuthExpired {
		definition.Auth.Token = "expired"
	}
	for _, event := range recording.Events {
		switch event.Kind {
		case domain.CaptureHTTPExchange:
			definition.Steps = append(definition.Steps, domain.ScenarioStep{
				Kind:  domain.StepWaitForHTTP,
				Value: fmt.Sprintf("%s %s %d", strings.ToUpper(event.HTTP.Method), event.HTTP.Path, event.HTTP.Status),
			})
			definition.Assertions = append(definition.Assertions,
				domain.ScenarioAssertion{Request: &domain.RequestExpectation{Method: strings.ToUpper(event.HTTP.Method), Path: event.HTTP.Path}},
				domain.ScenarioAssertion{Response: &domain.ResponseExpectation{Status: event.HTTP.Status}},
			)
		case domain.CaptureDeepLink:
			definition.Steps = append(definition.Steps, domain.ScenarioStep{Kind: domain.StepOpenDeepLink, Value: event.DeepLink.URL})
		case domain.CaptureEnvironment:
			mutation := event.Environment
			switch mutation.Action {
			case "latency":
				definition.Steps = append(definition.Steps, domain.ScenarioStep{Kind: domain.StepSetLatency, Value: strconv.Itoa(mutation.LatencyMS)})
			case "error":
				definition.Steps = append(definition.Steps, domain.ScenarioStep{Kind: domain.StepSetError, Value: strconv.Itoa(mutation.ForcedError)})
			case "auth":
				kind := domain.StepResetAuth
				if mutation.AuthExpired {
					kind = domain.StepExpireAuth
				}
				definition.Steps = append(definition.Steps, domain.ScenarioStep{Kind: kind})
			case "reset":
				definition.Steps = append(definition.Steps, domain.ScenarioStep{Kind: domain.StepResetAPI})
			default:
				return domain.ScenarioDefinition{}, fmt.Errorf("unsupported recorded environment action %q", mutation.Action)
			}
		}
	}
	if err := definition.Validate(); err != nil {
		return domain.ScenarioDefinition{}, fmt.Errorf("generated scenario is invalid: %w", err)
	}
	return definition, nil
}

type scenarioYAML struct {
	SchemaVersion int                       `yaml:"schema_version"`
	Name          string                    `yaml:"name"`
	Backend       *scenarioBackendYAML      `yaml:"backend,omitempty"`
	Auth          *scenarioAuthYAML         `yaml:"auth,omitempty"`
	Steps         []any                     `yaml:"steps,omitempty"`
	Expect        []scenarioExpectationYAML `yaml:"expect,omitempty"`
}

type scenarioBackendYAML struct {
	Latency int `yaml:"latency,omitempty"`
	Error   int `yaml:"error,omitempty"`
}
type scenarioAuthYAML struct {
	Token string `yaml:"token"`
}
type scenarioExpectationYAML struct {
	Request  *requestExpectationYAML  `yaml:"request,omitempty"`
	Response *responseExpectationYAML `yaml:"response,omitempty"`
}
type requestExpectationYAML struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}
type responseExpectationYAML struct {
	Status int `yaml:"status"`
}

func EncodeScenario(definition domain.ScenarioDefinition) ([]byte, error) {
	dto := scenarioYAML{SchemaVersion: schemas.ScenarioVersion, Name: definition.Name}
	if definition.Backend.LatencyMS != 0 || definition.Backend.Error != 0 {
		dto.Backend = &scenarioBackendYAML{Latency: definition.Backend.LatencyMS, Error: definition.Backend.Error}
	}
	if definition.Auth.Token != "" {
		dto.Auth = &scenarioAuthYAML{Token: definition.Auth.Token}
	}
	for _, step := range definition.Steps {
		if step.Value == "" {
			dto.Steps = append(dto.Steps, string(step.Kind))
		} else {
			dto.Steps = append(dto.Steps, map[string]string{string(step.Kind): step.Value})
		}
	}
	for _, assertion := range definition.Assertions {
		expectation := scenarioExpectationYAML{}
		if assertion.Request != nil {
			expectation.Request = &requestExpectationYAML{Method: assertion.Request.Method, Path: assertion.Request.Path}
		}
		if assertion.Response != nil {
			expectation.Response = &responseExpectationYAML{Status: assertion.Response.Status}
		}
		dto.Expect = append(dto.Expect, expectation)
	}
	data, err := yaml.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("encode scenario YAML: %w", err)
	}
	return data, nil
}

func WriteScenario(path string, data []byte, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("scenario %s already exists; use --force to replace it", filepath.Base(path))
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create scenario directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".mobilelab-record-*.yaml")
	if err != nil {
		return fmt.Errorf("create temporary scenario: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace scenario: %w", err)
	}
	return nil
}
