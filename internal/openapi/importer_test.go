package openapi

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/config"
)

func TestImportGeneratesDeterministicMocksAndScenarios(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, config.DefaultFilename)
	if err := config.Write(configPath, config.Default("openapi-test")); err != nil {
		t.Fatal(err)
	}
	specification := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /api/users/{id}:
    get:
      operationId: getUser
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        '200':
          description: User
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id: {type: string, example: '123'}
        active: {type: boolean}
`
	specificationPath := filepath.Join(root, "openapi.yaml")
	if err := os.WriteFile(specificationPath, []byte(specification), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Import(context.Background(), specificationPath, configPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.Endpoints != 1 || result.Schemas != 1 || result.ScenariosGenerated != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, endpoint := range cfg.Endpoints {
		if endpoint.Path == "/api/users/{id}" {
			found = true
			body := endpoint.Response.Body.(map[string]any)
			if body["id"] != "123" || body["active"] != false {
				t.Fatalf("unexpected generated body: %#v", body)
			}
		}
	}
	if !found {
		t.Fatal("generated endpoint missing from config")
	}
	scenarioData, err := os.ReadFile(filepath.Join(root, "mobilelab", "scenarios", "openapi-getuser.yaml"))
	if err != nil {
		t.Fatalf("generated scenario missing: %v", err)
	}
	if !strings.HasPrefix(string(scenarioData), "schema_version: 1\n") {
		t.Fatalf("generated scenario is not schema v1: %s", scenarioData)
	}
}
