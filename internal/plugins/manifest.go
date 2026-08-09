package plugins

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	protocol "github.com/mobilelab-dev/mobilelab/pkg/plugin"
	"gopkg.in/yaml.v3"
)

const (
	ManifestFilename = "plugin.yaml"
	MaxManifestBytes = 64 << 10
)

var (
	pluginNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)
	semverPattern     = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
)

type Manifest struct {
	APIVersion  string             `yaml:"api_version" json:"api_version"`
	Name        string             `yaml:"name" json:"name"`
	Version     string             `yaml:"version" json:"version"`
	Description string             `yaml:"description" json:"description"`
	Executable  string             `yaml:"executable" json:"executable"`
	Actions     []ActionDefinition `yaml:"actions" json:"actions"`
}

type ActionDefinition struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
}

func LoadManifest(path string) (Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read plugin manifest %q: %w", path, err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, MaxManifestBytes+1))
	if err != nil {
		return Manifest{}, fmt.Errorf("read plugin manifest %q: %w", path, err)
	}
	if len(data) > MaxManifestBytes {
		return Manifest{}, fmt.Errorf("plugin manifest %q exceeds %d bytes", path, MaxManifestBytes)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse plugin manifest %q: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Manifest{}, fmt.Errorf("parse plugin manifest %q: expected exactly one YAML document", path)
		}
		return Manifest{}, fmt.Errorf("parse plugin manifest %q: %w", path, err)
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, fmt.Errorf("invalid plugin manifest %q: %w", path, err)
	}
	return manifest, nil
}

func (manifest Manifest) Validate() error {
	if manifest.APIVersion != protocol.ProtocolVersion {
		return fmt.Errorf("api_version must be %q", protocol.ProtocolVersion)
	}
	if !pluginNamePattern.MatchString(manifest.Name) {
		return fmt.Errorf("name must match %s", pluginNamePattern)
	}
	if !semverPattern.MatchString(manifest.Version) {
		return fmt.Errorf("version must be valid Semantic Versioning")
	}
	if prerelease := semanticPrerelease(manifest.Version); prerelease != "" {
		for _, identifier := range strings.Split(prerelease, ".") {
			if len(identifier) > 1 && identifier[0] == '0' && onlyDigits(identifier) {
				return fmt.Errorf("version must be valid Semantic Versioning")
			}
		}
	}
	if strings.TrimSpace(manifest.Description) == "" || len(manifest.Description) > 256 {
		return fmt.Errorf("description is required and must not exceed 256 bytes")
	}
	if manifest.Executable == "" || filepath.Base(manifest.Executable) != manifest.Executable || manifest.Executable == "." || manifest.Executable == ".." {
		return fmt.Errorf("executable must be a file name inside the plugin directory")
	}
	if len(manifest.Actions) == 0 || len(manifest.Actions) > 32 {
		return fmt.Errorf("actions must contain between 1 and 32 entries")
	}
	seen := make(map[string]struct{}, len(manifest.Actions))
	for index, action := range manifest.Actions {
		if !pluginNamePattern.MatchString(action.Name) {
			return fmt.Errorf("actions[%d].name must match %s", index, pluginNamePattern)
		}
		if strings.TrimSpace(action.Description) == "" || len(action.Description) > 256 {
			return fmt.Errorf("actions[%d].description is required and must not exceed 256 bytes", index)
		}
		if _, exists := seen[action.Name]; exists {
			return fmt.Errorf("duplicate action %q", action.Name)
		}
		seen[action.Name] = struct{}{}
	}
	return nil
}

func semanticPrerelease(version string) string {
	start := strings.IndexByte(version, '-')
	if start < 0 {
		return ""
	}
	value := version[start+1:]
	if end := strings.IndexByte(value, '+'); end >= 0 {
		value = value[:end]
	}
	return value
}

func onlyDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func (manifest Manifest) Supports(action string) bool {
	for _, candidate := range manifest.Actions {
		if candidate.Name == action {
			return true
		}
	}
	return false
}
