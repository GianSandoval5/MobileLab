// Package datastore implements the optional, project-local business data API.
// It intentionally uses a database separate from MobileLab's internal history.
package datastore

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/schemas"
	"gopkg.in/yaml.v3"
)

const (
	ConfigFilename   = "data.yaml"
	DatabaseFilename = "data.db"
)

var resourceNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
var idFieldPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]{0,63}$`)

type Config struct {
	SchemaVersion int                           `yaml:"schema_version"`
	Resources     map[string]ResourceDefinition `yaml:"resources"`
}

type ResourceDefinition struct {
	Path      string `yaml:"path"`
	ID        string `yaml:"id,omitempty"`
	Singleton bool   `yaml:"singleton,omitempty"`
	Seed      string `yaml:"seed,omitempty"`
	Protected bool   `yaml:"protected,omitempty"`
}

func ConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, "mobilelab", ConfigFilename)
}

func DatabasePath(projectRoot string) string {
	return filepath.Join(projectRoot, "mobilelab", DatabaseFilename)
}

func LoadOptional(projectRoot string) (Config, bool, error) {
	path := ConfigPath(projectRoot)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, false, nil
	}
	if err != nil {
		return Config{}, false, fmt.Errorf("read data configuration %q: %w", path, err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, schemas.MaxYAMLBytes+1))
	if err != nil {
		return Config{}, false, fmt.Errorf("read data configuration %q: %w", path, err)
	}
	cfg, err := Parse(data)
	if err != nil {
		return Config{}, false, fmt.Errorf("invalid data configuration %q: %w", path, err)
	}
	return cfg, true, nil
}

func Parse(data []byte) (Config, error) {
	if len(data) > schemas.MaxYAMLBytes {
		return Config{}, fmt.Errorf("data configuration exceeds %d bytes", schemas.MaxYAMLBytes)
	}
	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, errors.New("expected exactly one YAML document")
		}
		return Config{}, fmt.Errorf("parse YAML: %w", err)
	}
	for name, resource := range cfg.Resources {
		if resource.ID == "" {
			resource.ID = "id"
			cfg.Resources[name] = resource
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var problems []error
	if c.SchemaVersion != schemas.DataVersion {
		problems = append(problems, fmt.Errorf("schema_version must be %d (got %d)", schemas.DataVersion, c.SchemaVersion))
	}
	if len(c.Resources) == 0 {
		problems = append(problems, errors.New("resources must contain at least one resource"))
	}
	paths := map[string]string{}
	for name, resource := range c.Resources {
		prefix := "resources." + name
		if !resourceNamePattern.MatchString(name) {
			problems = append(problems, fmt.Errorf("%s has an invalid name", prefix))
		}
		if !idFieldPattern.MatchString(resource.ID) {
			problems = append(problems, fmt.Errorf("%s.id must be a simple JSON field name", prefix))
		}
		if resource.Path == "/" || !strings.HasPrefix(resource.Path, "/") || path.Clean(resource.Path) != resource.Path || strings.HasSuffix(resource.Path, "/") || strings.ContainsAny(resource.Path, "{}?# \t\r\n") {
			problems = append(problems, fmt.Errorf("%s.path must be an absolute normalized non-root path without a trailing slash", prefix))
		}
		if owner, exists := paths[resource.Path]; exists {
			problems = append(problems, fmt.Errorf("%s.path conflicts with resources.%s", prefix, owner))
		}
		paths[resource.Path] = name
		if resource.Seed != "" {
			clean := filepath.Clean(resource.Seed)
			if filepath.IsAbs(resource.Seed) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
				problems = append(problems, fmt.Errorf("%s.seed must stay inside the mobilelab directory", prefix))
			}
		}
	}
	for firstName, first := range c.Resources {
		if first.Singleton {
			continue
		}
		for secondName, second := range c.Resources {
			if firstName == secondName || !strings.HasPrefix(second.Path, first.Path+"/") {
				continue
			}
			suffix := strings.TrimPrefix(second.Path, first.Path+"/")
			if suffix != "" && !strings.Contains(suffix, "/") {
				problems = append(problems, fmt.Errorf("resources.%s.path is owned by the item route of resources.%s", secondName, firstName))
			}
		}
	}
	return errors.Join(problems...)
}

func (c Config) Names() []string {
	names := make([]string, 0, len(c.Resources))
	for name := range c.Resources {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c Config) ValidateEndpoints(endpoints []config.EndpointDefinition) error {
	for _, endpoint := range endpoints {
		method := strings.ToUpper(endpoint.Method)
		for name, resource := range c.Resources {
			if endpoint.Path == resource.Path {
				if resource.Singleton && (method == "GET" || method == "PUT" || method == "PATCH") {
					return fmt.Errorf("endpoint %s %s conflicts with database resource %q", method, endpoint.Path, name)
				}
				if !resource.Singleton && (method == "GET" || method == "POST") {
					return fmt.Errorf("endpoint %s %s conflicts with database resource %q", method, endpoint.Path, name)
				}
			}
			if !resource.Singleton && strings.HasPrefix(endpoint.Path, resource.Path+"/") && strings.Count(strings.TrimPrefix(endpoint.Path, resource.Path+"/"), "/") == 0 && (method == "GET" || method == "PUT" || method == "PATCH" || method == "DELETE") {
				return fmt.Errorf("endpoint %s %s conflicts with database resource %q item route", method, endpoint.Path, name)
			}
		}
	}
	return nil
}

func ResolveSeed(workspace, relative string) (string, error) {
	if relative == "" {
		return "", nil
	}
	root, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		return "", fmt.Errorf("resolve mobilelab directory: %w", err)
	}
	target, err := filepath.EvalSymlinks(filepath.Join(workspace, relative))
	if err != nil {
		return "", fmt.Errorf("resolve seed %q: %w", relative, err)
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("seed %q escapes the mobilelab directory", relative)
	}
	return target, nil
}
