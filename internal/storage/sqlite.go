package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
	_ "modernc.org/sqlite"
)

const currentSchemaVersion = 3

type SQLite struct {
	db *sql.DB
}

func OpenSQLite(path string) (*SQLite, error) {
	if path == "" {
		return nil, fmt.Errorf("SQLite path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create SQLite directory: %w", err)
	}
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	database.SetMaxOpenConns(1)
	store := &SQLite{db: database}
	if err := store.initialize(context.Background()); err != nil {
		_ = database.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLite) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLite) initialize(ctx context.Context) error {
	for _, statement := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)`,
	} {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("initialize SQLite: %w", err)
		}
	}
	return s.migrate(ctx)
}

func (s *SQLite) migrate(ctx context.Context) error {
	var version int
	err := s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return fmt.Errorf("read SQLite schema version: %w", err)
	}
	if version > currentSchemaVersion {
		return fmt.Errorf("database schema version %d is newer than supported version %d", version, currentSchemaVersion)
	}
	for next := version + 1; next <= currentSchemaVersion; next++ {
		if err := s.applyMigration(ctx, next); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLite) applyMigration(ctx context.Context, version int) error {
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin SQLite migration %d: %w", version, err)
	}
	defer transaction.Rollback()

	statements, found := migrations[version]
	if !found {
		return fmt.Errorf("SQLite migration %d is not defined", version)
	}
	for _, statement := range statements {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply SQLite migration %d: %w", version, err)
		}
	}
	if _, err := transaction.ExecContext(ctx,
		"INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)",
		version, time.Now().UTC().Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("record SQLite migration %d: %w", version, err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit SQLite migration %d: %w", version, err)
	}
	return nil
}

var migrations = map[int][]string{
	1: {
		`CREATE TABLE request_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			method TEXT NOT NULL,
			path TEXT NOT NULL,
			query_json TEXT,
			headers_json TEXT,
			body_json TEXT,
			status INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			timestamp TEXT NOT NULL
		)`,
		`CREATE INDEX request_history_timestamp_idx ON request_history(timestamp DESC, id DESC)`,
		`CREATE TABLE scenario_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			passed INTEGER NOT NULL,
			started_at TEXT NOT NULL,
			duration_ms INTEGER NOT NULL,
			result_json TEXT NOT NULL
		)`,
		`CREATE INDEX scenario_runs_started_at_idx ON scenario_runs(started_at DESC, id DESC)`,
	},
	2: {
		`CREATE TABLE app_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			protocol_version INTEGER NOT NULL,
			framework TEXT NOT NULL,
			kind TEXT NOT NULL,
			name TEXT NOT NULL,
			passed INTEGER,
			session_id TEXT,
			attributes_json TEXT,
			timestamp TEXT NOT NULL
		)`,
		`CREATE INDEX app_events_timestamp_idx ON app_events(timestamp DESC, id DESC)`,
	},
	3: {
		`ALTER TABLE request_history ADD COLUMN response_headers_json TEXT`,
		`ALTER TABLE request_history ADD COLUMN response_body_json TEXT`,
	},
}

func (s *SQLite) Append(ctx context.Context, record domain.RequestRecord) error {
	query, err := marshalNullable(record.Query)
	if err != nil {
		return fmt.Errorf("encode request query: %w", err)
	}
	headers, err := marshalNullable(record.Headers)
	if err != nil {
		return fmt.Errorf("encode request headers: %w", err)
	}
	body, err := marshalNullable(record.Body)
	if err != nil {
		return fmt.Errorf("encode request body: %w", err)
	}
	responseHeaders, err := marshalNullable(record.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("encode response headers: %w", err)
	}
	responseBody, err := marshalNullable(record.ResponseBody)
	if err != nil {
		return fmt.Errorf("encode response body: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO request_history(
		method, path, query_json, headers_json, body_json, status, duration_ms, timestamp, response_headers_json, response_body_json
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.Method, record.Path, query, headers, body, record.Status, record.DurationMS, record.Timestamp.UTC().Format(time.RFC3339Nano), responseHeaders, responseBody,
	)
	if err != nil {
		return fmt.Errorf("store request record: %w", err)
	}
	return nil
}

func (s *SQLite) Recent(ctx context.Context, limit int) ([]domain.RequestRecord, error) {
	limit = normalizedLimit(limit)
	rows, err := s.db.QueryContext(ctx, `SELECT method, path, query_json, headers_json, body_json, status, duration_ms, timestamp, response_headers_json, response_body_json
		FROM (
			SELECT id, method, path, query_json, headers_json, body_json, status, duration_ms, timestamp, response_headers_json, response_body_json
			FROM request_history ORDER BY timestamp DESC, id DESC LIMIT ?
		) ORDER BY timestamp ASC, id ASC`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent requests: %w", err)
	}
	defer rows.Close()

	var records []domain.RequestRecord
	for rows.Next() {
		var record domain.RequestRecord
		var query, headers, body, responseHeaders, responseBody sql.NullString
		var timestamp string
		if err := rows.Scan(&record.Method, &record.Path, &query, &headers, &body, &record.Status, &record.DurationMS, &timestamp, &responseHeaders, &responseBody); err != nil {
			return nil, fmt.Errorf("scan request record: %w", err)
		}
		if err := unmarshalNullable(query, &record.Query); err != nil {
			return nil, fmt.Errorf("decode request query: %w", err)
		}
		if err := unmarshalNullable(headers, &record.Headers); err != nil {
			return nil, fmt.Errorf("decode request headers: %w", err)
		}
		if err := unmarshalNullable(body, &record.Body); err != nil {
			return nil, fmt.Errorf("decode request body: %w", err)
		}
		if err := unmarshalNullable(responseHeaders, &record.ResponseHeaders); err != nil {
			return nil, fmt.Errorf("decode response headers: %w", err)
		}
		if err := unmarshalNullable(responseBody, &record.ResponseBody); err != nil {
			return nil, fmt.Errorf("decode response body: %w", err)
		}
		record.Timestamp, err = time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			return nil, fmt.Errorf("decode request timestamp: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate request records: %w", err)
	}
	return records, nil
}

func (s *SQLite) Save(ctx context.Context, result domain.ScenarioResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode scenario result: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO scenario_runs(name, passed, started_at, duration_ms, result_json)
		VALUES(?, ?, ?, ?, ?)`,
		result.Name, result.Passed, result.StartedAt.UTC().Format(time.RFC3339Nano), result.DurationMS, string(data),
	)
	if err != nil {
		return fmt.Errorf("store scenario result: %w", err)
	}
	return nil
}

func (s *SQLite) RecentScenarioRuns(ctx context.Context, limit int) ([]domain.ScenarioResult, error) {
	limit = normalizedLimit(limit)
	rows, err := s.db.QueryContext(ctx, `SELECT result_json FROM (
		SELECT id, started_at, result_json FROM scenario_runs ORDER BY started_at DESC, id DESC LIMIT ?
	) ORDER BY started_at ASC, id ASC`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent scenario runs: %w", err)
	}
	defer rows.Close()
	var results []domain.ScenarioResult
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan scenario result: %w", err)
		}
		var result domain.ScenarioResult
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return nil, fmt.Errorf("decode scenario result: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scenario results: %w", err)
	}
	return results, nil
}

func (s *SQLite) SaveAppEvent(ctx context.Context, event domain.AppEvent) error {
	attributes, err := marshalNullable(event.Attributes)
	if err != nil {
		return fmt.Errorf("encode app event attributes: %w", err)
	}
	var passed any
	if event.Passed != nil {
		passed = *event.Passed
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO app_events(
		protocol_version, framework, kind, name, passed, session_id, attributes_json, timestamp
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ProtocolVersion, event.Framework, event.Kind, event.Name, passed, event.SessionID, attributes, event.Timestamp.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("store app event: %w", err)
	}
	return nil
}

func (s *SQLite) RecentAppEvents(ctx context.Context, limit int) ([]domain.AppEvent, error) {
	limit = normalizedLimit(limit)
	rows, err := s.db.QueryContext(ctx, `SELECT protocol_version, framework, kind, name, passed, session_id, attributes_json, timestamp FROM (
		SELECT id, protocol_version, framework, kind, name, passed, session_id, attributes_json, timestamp
		FROM app_events ORDER BY timestamp DESC, id DESC LIMIT ?
	) ORDER BY timestamp ASC, id ASC`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent app events: %w", err)
	}
	defer rows.Close()
	var events []domain.AppEvent
	for rows.Next() {
		var event domain.AppEvent
		var passed sql.NullBool
		var sessionID, attributes sql.NullString
		var timestamp string
		if err := rows.Scan(&event.ProtocolVersion, &event.Framework, &event.Kind, &event.Name, &passed, &sessionID, &attributes, &timestamp); err != nil {
			return nil, fmt.Errorf("scan app event: %w", err)
		}
		if passed.Valid {
			value := passed.Bool
			event.Passed = &value
		}
		if sessionID.Valid {
			event.SessionID = sessionID.String
		}
		if err := unmarshalNullable(attributes, &event.Attributes); err != nil {
			return nil, fmt.Errorf("decode app event attributes: %w", err)
		}
		event.Timestamp, err = time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			return nil, fmt.Errorf("decode app event timestamp: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate app events: %w", err)
	}
	return events, nil
}

// Recent implements domain.ScenarioRunRepository through a named adapter method.
// Go cannot overload Recent for both repository interfaces, so ScenarioRuns exposes
// a small wrapper that preserves the domain ports without leaking SQL.
type ScenarioRuns struct{ store *SQLite }

func (s *SQLite) ScenarioRuns() ScenarioRuns { return ScenarioRuns{store: s} }

func (r ScenarioRuns) Save(ctx context.Context, result domain.ScenarioResult) error {
	return r.store.Save(ctx, result)
}

func (r ScenarioRuns) Recent(ctx context.Context, limit int) ([]domain.ScenarioResult, error) {
	return r.store.RecentScenarioRuns(ctx, limit)
}

func marshalNullable(value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func unmarshalNullable(value sql.NullString, destination any) error {
	if !value.Valid {
		return nil
	}
	return json.Unmarshal([]byte(value.String), destination)
}

func normalizedLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

var (
	_ domain.RequestRepository     = (*SQLite)(nil)
	_ domain.ScenarioRunRepository = ScenarioRuns{}
	_ domain.AppEventRepository    = (*SQLite)(nil)
)
