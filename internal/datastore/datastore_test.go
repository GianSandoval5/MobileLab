package datastore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mobilelab-dev/mobilelab/internal/config"
)

func TestParseIsStrictAndValidatesResources(t *testing.T) {
	valid := []byte("schema_version: 1\nresources:\n  products:\n    path: /api/products\n    seed: seeds/products.json\n")
	cfg, err := Parse(valid)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Resources["products"].ID != "id" {
		t.Fatalf("default id field was not applied: %#v", cfg)
	}
	for _, invalid := range []string{
		"schema_version: 2\nresources:\n  products:\n    path: /api/products\n",
		"schema_version: 1\nresources:\n  Products:\n    path: /api/products\n",
		"schema_version: 1\nresources:\n  products:\n    path: /api/products/\n",
		"schema_version: 1\nresources:\n  products:\n    path: /api/products\n    surprise: true\n",
		"schema_version: 1\nresources:\n  products:\n    path: /api/products\n    seed: ../outside.json\n",
		"schema_version: 1\nresources:\n  products:\n    path: /api/products\n  featured:\n    path: /api/products/featured\n",
	} {
		if _, err := Parse([]byte(invalid)); err == nil {
			t.Fatalf("invalid data configuration accepted:\n%s", invalid)
		}
	}
}

func TestValidateEndpointsRejectsOwnedRoutes(t *testing.T) {
	cfg, err := Parse([]byte("schema_version: 1\nresources:\n  products:\n    path: /api/products\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = cfg.ValidateEndpoints([]config.EndpointDefinition{{Method: http.MethodGet, Path: "/api/products"}})
	if err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("expected route conflict, got %v", err)
	}
}

func TestSeedAndCRUDHandlerPersistDocuments(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "mobilelab")
	if err := os.MkdirAll(filepath.Join(workspace, "seeds"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "seeds", "products.json"), []byte(`[{"id":"one","name":"First"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "seeds", "profile.json"), []byte(`{"id":"user","name":"Demo"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse([]byte("schema_version: 1\nresources:\n  products:\n    path: /api/products\n    seed: seeds/products.json\n  profile:\n    path: /api/profile\n    singleton: true\n    seed: seeds/profile.json\n"))
	if err != nil {
		t.Fatal(err)
	}
	store, err := Open(DatabasePath(root))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Seed(context.Background(), cfg, workspace, SeedEmpty); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(cfg, store)
	if err != nil {
		t.Fatal(err)
	}

	assertResponse(t, handler, http.MethodGet, "/api/products", "", http.StatusOK, `[{"id":"one","name":"First"}]`)
	created := assertResponse(t, handler, http.MethodPost, "/api/products", `{"name":"Second"}`, http.StatusCreated, "")
	var product map[string]any
	if err := json.Unmarshal(created.Body.Bytes(), &product); err != nil || product["id"] == "" {
		t.Fatalf("generated product is invalid: %s (%v)", created.Body, err)
	}
	id := product["id"].(string)
	assertResponse(t, handler, http.MethodPatch, "/api/products/"+id, `{"name":"Updated"}`, http.StatusOK, "")
	assertResponse(t, handler, http.MethodGet, "/api/products/"+id, "", http.StatusOK, `{"id":"`+id+`","name":"Updated"}`)
	assertResponse(t, handler, http.MethodDelete, "/api/products/"+id, "", http.StatusNoContent, "")
	assertResponse(t, handler, http.MethodGet, "/api/products/"+id, "", http.StatusNotFound, "")
	assertResponse(t, handler, http.MethodPatch, "/api/profile", `{"name":"Changed"}`, http.StatusOK, `{"id":"user","name":"Changed"}`)

	// Empty seeding is non-destructive; reset is intentionally destructive.
	if err := store.Seed(context.Background(), cfg, workspace, SeedEmpty); err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodGet, "/api/profile", "", http.StatusOK, `{"id":"user","name":"Changed"}`)
	if err := store.Seed(context.Background(), cfg, workspace, SeedReset); err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodGet, "/api/profile", "", http.StatusOK, `{"id":"user","name":"Demo"}`)
}

func TestResolveSeedRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "mobilelab")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside.json")
	if err := os.WriteFile(outside, []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(workspace, "seed.json")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := ResolveSeed(workspace, "seed.json"); err == nil {
		t.Fatal("seed symlink escaped the workspace")
	}
}

func TestOpenRejectsNewerBusinessDatabaseSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`UPDATE metadata SET value = '2' WHERE key = 'schema_version'`); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("newer database schema was accepted: %v", err)
	}
}

func assertResponse(t *testing.T, handler http.Handler, method, path, body string, status int, expectedJSON string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != status {
		t.Fatalf("%s %s returned %d: %s", method, path, response.Code, response.Body.String())
	}
	if expectedJSON != "" {
		var got, want any
		if json.Unmarshal(response.Body.Bytes(), &got) != nil || json.Unmarshal([]byte(expectedJSON), &want) != nil {
			t.Fatalf("invalid JSON comparison: got=%s want=%s", response.Body, expectedJSON)
		}
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(want)
		if string(gotJSON) != string(wantJSON) {
			t.Fatalf("unexpected JSON: got=%s want=%s", gotJSON, wantJSON)
		}
	}
	return response
}
