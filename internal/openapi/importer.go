package openapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/schemas"
	"gopkg.in/yaml.v3"
)

type Result struct {
	Endpoints          int
	Schemas            int
	ScenariosGenerated int
}

type generatedEndpoint struct {
	definition  config.EndpointDefinition
	operationID string
}

type generatedScenario struct {
	SchemaVersion int    `yaml:"schema_version"`
	Name          string `yaml:"name"`
	Expect        []any  `yaml:"expect"`
}

func Import(ctx context.Context, specificationPath, configPath string) (Result, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	document, err := loader.LoadFromFile(specificationPath)
	if err != nil {
		return Result{}, fmt.Errorf("load OpenAPI document: %w", err)
	}
	if err := document.Validate(ctx); err != nil {
		return Result{}, fmt.Errorf("validate OpenAPI document: %w", err)
	}
	generated := generateEndpoints(document)
	if len(generated) == 0 {
		return Result{}, fmt.Errorf("OpenAPI document does not contain operations")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return Result{}, err
	}
	byRoute := make(map[string]int, len(cfg.Endpoints))
	for index, endpoint := range cfg.Endpoints {
		byRoute[strings.ToUpper(endpoint.Method)+" "+endpoint.Path] = index
	}
	for _, endpoint := range generated {
		key := endpoint.definition.Method + " " + endpoint.definition.Path
		if index, exists := byRoute[key]; exists {
			cfg.Endpoints[index] = endpoint.definition
		} else {
			byRoute[key] = len(cfg.Endpoints)
			cfg.Endpoints = append(cfg.Endpoints, endpoint.definition)
		}
	}
	if err := config.Write(configPath, cfg); err != nil {
		return Result{}, err
	}

	scenarios, err := writeScenarios(filepath.Join(filepath.Dir(configPath), "mobilelab", "scenarios"), generated)
	if err != nil {
		return Result{}, err
	}
	schemaCount := 0
	if document.Components != nil {
		schemaCount = len(document.Components.Schemas)
	}
	return Result{Endpoints: len(generated), Schemas: schemaCount, ScenariosGenerated: scenarios}, nil
}

func generateEndpoints(document *openapi3.T) []generatedEndpoint {
	paths := document.Paths.InMatchingOrder()
	var result []generatedEndpoint
	for _, path := range paths {
		item := document.Paths.Value(path)
		methods := make([]string, 0, len(item.Operations()))
		for method := range item.Operations() {
			methods = append(methods, method)
		}
		sort.Strings(methods)
		for _, method := range methods {
			operation := item.Operations()[method]
			status, body := exampleResponse(operation)
			result = append(result, generatedEndpoint{
				definition: config.EndpointDefinition{
					Path: path, Method: strings.ToUpper(method),
					Response: config.EndpointResponse{Status: status, Body: body},
				},
				operationID: operation.OperationID,
			})
		}
	}
	return result
}

func exampleResponse(operation *openapi3.Operation) (int, any) {
	if operation == nil || operation.Responses == nil {
		return http.StatusOK, map[string]any{}
	}
	responses := operation.Responses.Map()
	keys := make([]string, 0, len(responses))
	for key := range responses {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	selectedStatus := http.StatusOK
	var selected *openapi3.ResponseRef
	for _, key := range keys {
		status, err := strconv.Atoi(key)
		if err == nil && status >= 200 && status <= 299 {
			selectedStatus, selected = status, responses[key]
			break
		}
	}
	if selected == nil {
		selected = responses["default"]
	}
	if selected == nil || selected.Value == nil || len(selected.Value.Content) == 0 {
		return selectedStatus, map[string]any{}
	}
	media := selected.Value.Content.Get("application/json")
	if media == nil {
		mediaTypes := make([]string, 0, len(selected.Value.Content))
		for mediaType := range selected.Value.Content {
			mediaTypes = append(mediaTypes, mediaType)
		}
		sort.Strings(mediaTypes)
		media = selected.Value.Content[mediaTypes[0]]
	}
	if media.Example != nil {
		return selectedStatus, media.Example
	}
	if media.Schema == nil {
		return selectedStatus, map[string]any{}
	}
	return selectedStatus, schemaExample(media.Schema, make(map[*openapi3.Schema]bool), 0)
}

func schemaExample(reference *openapi3.SchemaRef, seen map[*openapi3.Schema]bool, depth int) any {
	if reference == nil || reference.Value == nil || depth > 6 {
		return nil
	}
	schema := reference.Value
	if schema.Example != nil {
		return schema.Example
	}
	if schema.Default != nil {
		return schema.Default
	}
	if len(schema.Enum) > 0 {
		return schema.Enum[0]
	}
	if seen[schema] {
		return nil
	}
	seen[schema] = true
	defer delete(seen, schema)
	if schema.Type != nil && schema.Type.Includes("object") || len(schema.Properties) > 0 {
		value := make(map[string]any, len(schema.Properties))
		keys := make([]string, 0, len(schema.Properties))
		for key := range schema.Properties {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value[key] = schemaExample(schema.Properties[key], seen, depth+1)
		}
		return value
	}
	if schema.Type != nil && schema.Type.Includes("array") {
		return []any{schemaExample(schema.Items, seen, depth+1)}
	}
	if schema.Type != nil && schema.Type.Includes("integer") {
		return 0
	}
	if schema.Type != nil && schema.Type.Includes("number") {
		return 0.0
	}
	if schema.Type != nil && schema.Type.Includes("boolean") {
		return false
	}
	return "string"
}

func writeScenarios(directory string, endpoints []generatedEndpoint) (int, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return 0, fmt.Errorf("create scenario directory: %w", err)
	}
	created := 0
	for _, endpoint := range endpoints {
		name := slug(endpoint.operationID)
		if name == "" {
			name = slug(strings.ToLower(endpoint.definition.Method) + "-" + endpoint.definition.Path)
		}
		path := filepath.Join(directory, "openapi-"+name+".yaml")
		payload := generatedScenario{
			SchemaVersion: schemas.ScenarioVersion,
			Name:          "OpenAPI " + endpoint.definition.Method + " " + endpoint.definition.Path,
			Expect: []any{
				map[string]any{"request": map[string]any{"method": endpoint.definition.Method, "path": endpoint.definition.Path}},
				map[string]any{"response": map[string]any{"status": endpoint.definition.Response.Status}},
			},
		}
		data, err := yaml.Marshal(payload)
		if err != nil {
			return created, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return created, fmt.Errorf("create generated scenario: %w", err)
		}
		if _, err := file.Write(data); err != nil {
			_ = file.Close()
			return created, fmt.Errorf("write generated scenario: %w", err)
		}
		if err := file.Close(); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func slug(value string) string {
	var result strings.Builder
	separator := false
	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			if separator && result.Len() > 0 {
				result.WriteByte('-')
			}
			separator = false
			result.WriteRune(char)
		} else {
			separator = true
		}
	}
	return strings.Trim(result.String(), "-")
}
