package scenario

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/schemas"
	"gopkg.in/yaml.v3"
)

type YAMLParser struct{}

type scenarioDTO struct {
	SchemaVersion int    `yaml:"schema_version,omitempty"`
	Name          string `yaml:"name"`
	Backend       struct {
		Latency int `yaml:"latency"`
		Error   int `yaml:"error"`
	} `yaml:"backend,omitempty"`
	Auth struct {
		Token string `yaml:"token"`
	} `yaml:"auth,omitempty"`
	Device struct {
		Network string `yaml:"network"`
	} `yaml:"device,omitempty"`
	Steps  []stepDTO      `yaml:"steps,omitempty"`
	Expect []assertionDTO `yaml:"expect,omitempty"`
}

type stepDTO struct {
	Kind  string
	Value string
}

func (s *stepDTO) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		s.Kind = node.Value
		return nil
	case yaml.MappingNode:
		if len(node.Content) != 2 || node.Content[0].Kind != yaml.ScalarNode || node.Content[1].Kind != yaml.ScalarNode {
			return fmt.Errorf("scenario step must contain exactly one string operation")
		}
		s.Kind = node.Content[0].Value
		s.Value = node.Content[1].Value
		return nil
	default:
		return fmt.Errorf("scenario step must be a command or single-key mapping")
	}
}

type assertionDTO struct {
	Request  *requestDTO  `yaml:"request,omitempty"`
	Response *responseDTO `yaml:"response,omitempty"`
	AppEvent *appEventDTO `yaml:"app_event,omitempty"`
}

type appEventDTO struct {
	Framework string `yaml:"framework,omitempty"`
	Kind      string `yaml:"kind"`
	Name      string `yaml:"name"`
	Passed    *bool  `yaml:"passed,omitempty"`
}

type requestDTO struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type responseDTO struct {
	Status int `yaml:"status"`
}

func (YAMLParser) Parse(data []byte) (domain.ScenarioDefinition, error) {
	if len(data) > schemas.MaxYAMLBytes {
		return domain.ScenarioDefinition{}, fmt.Errorf("parse scenario YAML: document exceeds %d bytes", schemas.MaxYAMLBytes)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var dto scenarioDTO
	if err := decoder.Decode(&dto); err != nil {
		return domain.ScenarioDefinition{}, fmt.Errorf("parse scenario YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return domain.ScenarioDefinition{}, fmt.Errorf("parse scenario YAML: expected exactly one YAML document")
		}
		return domain.ScenarioDefinition{}, fmt.Errorf("parse scenario YAML: %w", err)
	}
	if dto.SchemaVersion < 0 {
		return domain.ScenarioDefinition{}, fmt.Errorf("invalid scenario: schema_version cannot be negative")
	}
	if dto.SchemaVersion > schemas.ScenarioVersion {
		return domain.ScenarioDefinition{}, fmt.Errorf("invalid scenario: schema_version %d is newer than supported version %d; upgrade MobileLab", dto.SchemaVersion, schemas.ScenarioVersion)
	}
	definition := domain.ScenarioDefinition{
		Name:    dto.Name,
		Backend: domain.ScenarioBackend{LatencyMS: dto.Backend.Latency, Error: dto.Backend.Error},
		Auth:    domain.ScenarioAuth{Token: strings.ToLower(dto.Auth.Token)},
		Device:  domain.ScenarioDevice{Network: domain.NetworkCondition(strings.ToLower(dto.Device.Network))},
	}
	for _, step := range dto.Steps {
		definition.Steps = append(definition.Steps, domain.ScenarioStep{Kind: domain.ScenarioStepKind(step.Kind), Value: step.Value})
	}
	for _, assertion := range dto.Expect {
		converted := domain.ScenarioAssertion{}
		if assertion.Request != nil {
			converted.Request = &domain.RequestExpectation{Method: strings.ToUpper(assertion.Request.Method), Path: assertion.Request.Path}
		}
		if assertion.Response != nil {
			converted.Response = &domain.ResponseExpectation{Status: assertion.Response.Status}
		}
		if assertion.AppEvent != nil {
			converted.AppEvent = &domain.AppEventExpectation{
				Framework: domain.AppFramework(strings.ToLower(assertion.AppEvent.Framework)),
				Kind:      domain.AppEventKind(strings.ToLower(assertion.AppEvent.Kind)),
				Name:      assertion.AppEvent.Name,
				Passed:    assertion.AppEvent.Passed,
			}
		}
		definition.Assertions = append(definition.Assertions, converted)
	}
	if err := definition.Validate(); err != nil {
		return domain.ScenarioDefinition{}, fmt.Errorf("invalid scenario: %w", err)
	}
	return definition, nil
}
