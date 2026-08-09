package migration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/scenario"
	"github.com/mobilelab-dev/mobilelab/schemas"
	"gopkg.in/yaml.v3"
)

type Document struct {
	Path           string
	Kind           schemas.Kind
	FromVersion    int
	ToVersion      int
	NeedsMigration bool
	data           []byte
	mode           os.FileMode
}

type Plan struct {
	Documents []Document
}

func Build(projectDir string) (Plan, error) {
	if strings.TrimSpace(projectDir) == "" {
		return Plan{}, fmt.Errorf("project directory is required")
	}
	configPath := filepath.Join(projectDir, config.DefaultFilename)
	configuration, err := prepare(configPath, schemas.Config, schemas.ConfigVersion, func(data []byte) error {
		_, err := config.Parse(data)
		return err
	})
	if err != nil {
		return Plan{}, err
	}
	documents := []Document{configuration}
	scenarioDirectory := filepath.Join(projectDir, "mobilelab", "scenarios")
	err = filepath.WalkDir(scenarioDirectory, func(path string, entry os.DirEntry, walkErr error) error {
		if errors.Is(walkErr, os.ErrNotExist) && path == scenarioDirectory {
			return filepath.SkipDir
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension != ".yaml" && extension != ".yml" {
			return nil
		}
		document, err := prepare(path, schemas.Scenario, schemas.ScenarioVersion, func(data []byte) error {
			_, err := (scenario.YAMLParser{}).Parse(data)
			return err
		})
		if err != nil {
			return err
		}
		documents = append(documents, document)
		return nil
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Plan{}, fmt.Errorf("discover scenario migrations: %w", err)
	}
	sort.Slice(documents, func(i, j int) bool { return documents[i].Path < documents[j].Path })
	return Plan{Documents: documents}, nil
}

func (plan Plan) ChangeCount() int {
	count := 0
	for _, document := range plan.Documents {
		if document.NeedsMigration {
			count++
		}
	}
	return count
}

func (plan Plan) Apply() error {
	for _, document := range plan.Documents {
		if !document.NeedsMigration {
			continue
		}
		if err := writeAtomic(document.Path, document.data, document.mode); err != nil {
			return fmt.Errorf("migrate %s %q: %w", document.Kind, document.Path, err)
		}
	}
	return nil
}

func prepare(path string, kind schemas.Kind, currentVersion int, validate func([]byte) error) (Document, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Document{}, fmt.Errorf("inspect %s %q: %w", kind, path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Document{}, fmt.Errorf("%s %q must be a regular file, not a symlink", kind, path)
	}
	file, err := os.Open(path)
	if err != nil {
		return Document{}, fmt.Errorf("read %s %q: %w", kind, path, err)
	}
	data, readErr := io.ReadAll(io.LimitReader(file, schemas.MaxYAMLBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return Document{}, fmt.Errorf("read %s %q: %w", kind, path, readErr)
	}
	if closeErr != nil {
		return Document{}, fmt.Errorf("close %s %q: %w", kind, path, closeErr)
	}
	if len(data) > schemas.MaxYAMLBytes {
		return Document{}, fmt.Errorf("%s %q exceeds %d bytes", kind, path, schemas.MaxYAMLBytes)
	}
	if err := validate(data); err != nil {
		return Document{}, fmt.Errorf("preflight %s %q: %w", kind, path, err)
	}

	var root yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&root); err != nil {
		return Document{}, fmt.Errorf("decode %s %q for migration: %w", kind, path, err)
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return Document{}, fmt.Errorf("%s %q must contain one top-level mapping", kind, path)
	}
	mapping := root.Content[0]
	version := 0
	var versionNode *yaml.Node
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == "schema_version" {
			versionNode = mapping.Content[index+1]
			if err := versionNode.Decode(&version); err != nil {
				return Document{}, fmt.Errorf("%s %q schema_version must be an integer", kind, path)
			}
			break
		}
	}
	document := Document{Path: path, Kind: kind, FromVersion: version, ToVersion: currentVersion, mode: info.Mode().Perm()}
	if version == currentVersion {
		return document, nil
	}
	if version != 0 {
		return Document{}, fmt.Errorf("%s %q uses unsupported schema_version %d", kind, path, version)
	}
	document.NeedsMigration = true
	if versionNode == nil {
		key := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "schema_version"}
		value := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", currentVersion)}
		mapping.Content = append([]*yaml.Node{key, value}, mapping.Content...)
	} else {
		versionNode.Kind = yaml.ScalarNode
		versionNode.Tag = "!!int"
		versionNode.Value = fmt.Sprintf("%d", currentVersion)
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		return Document{}, fmt.Errorf("encode migrated %s %q: %w", kind, path, err)
	}
	if err := encoder.Close(); err != nil {
		return Document{}, fmt.Errorf("finish migrated %s %q: %w", kind, path, err)
	}
	document.data = output.Bytes()
	return document, nil
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".mobilelab-migrate-*.yaml")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
