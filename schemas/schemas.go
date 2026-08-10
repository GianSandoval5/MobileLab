// Package schemas exposes the stable MobileLab v1 configuration and scenario
// JSON Schemas. The files are embedded so released binaries can print the same
// contracts that are versioned in the repository.
package schemas

import (
	"embed"
	"fmt"
)

const (
	ConfigVersion   = 1
	ScenarioVersion = 1
	DataVersion     = 1
	MaxYAMLBytes    = 1 << 20
)

type Kind string

const (
	Config   Kind = "config"
	Scenario Kind = "scenario"
	Data     Kind = "data"
)

//go:embed *.schema.json
var files embed.FS

func Read(kind Kind) ([]byte, error) {
	name := ""
	switch kind {
	case Config:
		name = "mobilelab-config-v1.schema.json"
	case Scenario:
		name = "mobilelab-scenario-v1.schema.json"
	case Data:
		name = "mobilelab-data-v1.schema.json"
	default:
		return nil, fmt.Errorf("unknown MobileLab schema %q", kind)
	}
	data, err := files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read embedded %s schema: %w", kind, err)
	}
	return data, nil
}
