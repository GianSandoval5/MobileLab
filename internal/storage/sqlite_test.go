package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

func TestSQLitePersistsSanitizedRequestsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mobilelab.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	records := []domain.RequestRecord{
		{Method: "GET", Path: "/first", Status: 200, DurationMS: 2, Timestamp: now, Headers: map[string][]string{"Authorization": {"[REDACTED]"}}},
		{Method: "POST", Path: "/second", Status: 500, DurationMS: 8, Timestamp: now.Add(time.Second), Body: map[string]any{"password": "[REDACTED]"}},
	}
	for _, record := range records {
		if err := store.Append(context.Background(), record); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	recent, err := store.Recent(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 2 || recent[0].Path != "/first" || recent[1].Path != "/second" {
		t.Fatalf("unexpected records: %#v", recent)
	}
	if recent[0].Headers["Authorization"][0] != "[REDACTED]" {
		t.Fatalf("sanitized header changed: %#v", recent[0].Headers)
	}
	body := recent[1].Body.(map[string]any)
	if body["password"] != "[REDACTED]" {
		t.Fatalf("sanitized body changed: %#v", body)
	}
	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("database was not persisted: %v, %#v", err, info)
	}
}

func TestSQLiteReturnsOnlyMostRecentRecordsInChronologicalOrder(t *testing.T) {
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "mobilelab.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	base := time.Now().UTC()
	for index, path := range []string{"/one", "/two", "/three"} {
		if err := store.Append(context.Background(), domain.RequestRecord{Method: "GET", Path: path, Status: 200, Timestamp: base.Add(time.Duration(index) * time.Second)}); err != nil {
			t.Fatal(err)
		}
	}
	recent, err := store.Recent(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 2 || recent[0].Path != "/two" || recent[1].Path != "/three" {
		t.Fatalf("unexpected bounded history: %#v", recent)
	}
}

func TestSQLitePersistsScenarioResults(t *testing.T) {
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "mobilelab.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	repository := store.ScenarioRuns()
	want := domain.ScenarioResult{
		Name: "Payment failure", Passed: true, StartedAt: time.Now().UTC(), DurationMS: 42,
		Assertions: []domain.ScenarioCheck{{Name: "HTTP 500 observed", Passed: true}},
	}
	if err := repository.Save(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	results, err := repository.Recent(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != want.Name || !results[0].Passed || results[0].Assertions[0].Name != "HTTP 500 observed" {
		t.Fatalf("unexpected scenario history: %#v", results)
	}
}

func TestSQLiteRejectsNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mobilelab.db")
	store, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec("INSERT INTO schema_migrations(version, applied_at) VALUES(999, ?)", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = OpenSQLite(path)
	if err == nil || !strings.Contains(err.Error(), "newer than supported") {
		t.Fatalf("expected newer schema error, got %v", err)
	}
}
