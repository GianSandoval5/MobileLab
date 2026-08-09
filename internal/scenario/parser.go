package scenario

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"gopkg.in/yaml.v3"
)

type YAMLParser struct{}

type scenarioDTO struct {
	Name    string `yaml:"name"`
	Backend struct {
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
}

type requestDTO struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type responseDTO struct {
	Status int `yaml:"status"`
}

func (YAMLParser) Parse(data []byte) (domain.ScenarioDefinition, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var dto scenarioDTO
	if err := decoder.Decode(&dto); err != nil {
		return domain.ScenarioDefinition{}, fmt.Errorf("parse scenario YAML: %w", err)
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
		definition.Assertions = append(definition.Assertions, converted)
	}
	if err := definition.Validate(); err != nil {
		return domain.ScenarioDefinition{}, fmt.Errorf("invalid scenario: %w", err)
	}
	return definition, nil
}
