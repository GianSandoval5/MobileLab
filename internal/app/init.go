package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/detect"
	"github.com/mobilelab-dev/mobilelab/schemas"
)

type InitResult struct {
	ConfigPath string
	Created    []string
	Detected   []detect.Result
}

func Initialize(root string) (InitResult, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return InitResult{}, fmt.Errorf("resolve project directory: %w", err)
	}
	detected, err := detect.Project(absoluteRoot)
	if err != nil {
		return InitResult{}, fmt.Errorf("analyze project: %w", err)
	}

	configPath := filepath.Join(absoluteRoot, config.DefaultFilename)
	if _, err := os.Stat(configPath); err == nil {
		return InitResult{}, fmt.Errorf("%s already exists; MobileLab did not overwrite it", config.DefaultFilename)
	} else if !errors.Is(err, os.ErrNotExist) {
		return InitResult{}, fmt.Errorf("inspect %s: %w", config.DefaultFilename, err)
	}

	workspace := filepath.Join(absoluteRoot, "mobilelab")
	directories := []string{
		filepath.Join(workspace, "fixtures"),
		filepath.Join(workspace, "mocks"),
		filepath.Join(workspace, "scenarios"),
	}
	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return InitResult{}, fmt.Errorf("create %s: %w", directory, err)
		}
	}

	projectName := filepath.Base(absoluteRoot)
	if err := config.Write(configPath, config.Default(projectName)); err != nil {
		return InitResult{}, err
	}

	files := map[string]string{
		filepath.Join(workspace, ".gitignore"): `.state.json
*.db
*.db-*
`,
		filepath.Join(workspace, "fixtures", "profile.json"): `{
  "id": "{{userId}}",
  "name": "MobileLab User",
  "plan": "developer"
}
`,
		filepath.Join(workspace, "scenarios", "profile.yaml"): fmt.Sprintf(`schema_version: %d
name: Load local profile
steps:
  - launch_app
expect:
  - request:
      method: GET
      path: /api/profile
  - response:
      status: 200
`, schemas.ScenarioVersion),
	}
	created := []string{config.DefaultFilename, "mobilelab/fixtures/", "mobilelab/mocks/", "mobilelab/scenarios/"}
	for path, contents := range files {
		if err := writeNewFile(path, []byte(contents)); err != nil {
			return InitResult{}, err
		}
	}

	return InitResult{ConfigPath: configPath, Created: created, Detected: detected}, nil
}

func writeNewFile(path string, contents []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return nil
}
