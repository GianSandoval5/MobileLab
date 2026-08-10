package datastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const singletonID = "_singleton"

type Store struct {
	db *sql.DB
}

type SeedMode int

const (
	SeedEmpty SeedMode = iota
	SeedUpsert
	SeedReset
)

func Open(databasePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open business database: %w", err)
	}
	store := &Store{db: db}
	if err := store.initialize(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	if err := os.Chmod(databasePath, 0o600); err != nil {
		db.Close()
		return nil, fmt.Errorf("secure business database: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) initialize(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA busy_timeout=5000`,
		`CREATE TABLE IF NOT EXISTS metadata (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
		`INSERT INTO metadata(key, value) VALUES ('schema_version', '1') ON CONFLICT(key) DO NOTHING`,
		`CREATE TABLE IF NOT EXISTS documents (
            sequence INTEGER PRIMARY KEY AUTOINCREMENT,
            resource TEXT NOT NULL,
            document_id TEXT NOT NULL,
            document_json TEXT NOT NULL,
            created_at TEXT NOT NULL,
            updated_at TEXT NOT NULL,
            UNIQUE(resource, document_id)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_documents_resource_sequence ON documents(resource, sequence)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("initialize business database: %w", err)
		}
	}
	var schemaVersion string
	if err := s.db.QueryRowContext(ctx, `SELECT value FROM metadata WHERE key = 'schema_version'`).Scan(&schemaVersion); err != nil {
		return fmt.Errorf("read business database schema version: %w", err)
	}
	if schemaVersion != "1" {
		return fmt.Errorf("business database schema version %q is unsupported; upgrade MobileLab or restore a compatible data.db", schemaVersion)
	}
	return nil
}

func (s *Store) List(ctx context.Context, resource string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT document_json FROM documents WHERE resource = ? ORDER BY sequence`, resource)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", resource, err)
	}
	defer rows.Close()
	documents := make([]map[string]any, 0)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("read %s: %w", resource, err)
		}
		var document map[string]any
		if err := json.Unmarshal([]byte(raw), &document); err != nil {
			return nil, fmt.Errorf("decode stored %s document: %w", resource, err)
		}
		documents = append(documents, document)
	}
	return documents, rows.Err()
}

func (s *Store) Get(ctx context.Context, resource, id string) (map[string]any, bool, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT document_json FROM documents WHERE resource = ? AND document_id = ?`, resource, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get %s/%s: %w", resource, id, err)
	}
	var document map[string]any
	if err := json.Unmarshal([]byte(raw), &document); err != nil {
		return nil, false, fmt.Errorf("decode stored %s/%s: %w", resource, id, err)
	}
	return document, true, nil
}

func (s *Store) Create(ctx context.Context, resource, id string, document map[string]any) error {
	raw, err := json.Marshal(document)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(ctx, `INSERT INTO documents(resource, document_id, document_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, resource, id, raw, now, now)
	if err != nil {
		return fmt.Errorf("create %s/%s: %w", resource, id, err)
	}
	return nil
}

func (s *Store) Put(ctx context.Context, resource, id string, document map[string]any) error {
	raw, err := json.Marshal(document)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(ctx, `INSERT INTO documents(resource, document_id, document_json, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(resource, document_id) DO UPDATE SET document_json = excluded.document_json, updated_at = excluded.updated_at`, resource, id, raw, now, now)
	if err != nil {
		return fmt.Errorf("store %s/%s: %w", resource, id, err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, resource, id string) (bool, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM documents WHERE resource = ? AND document_id = ?`, resource, id)
	if err != nil {
		return false, fmt.Errorf("delete %s/%s: %w", resource, id, err)
	}
	rows, err := result.RowsAffected()
	return rows > 0, err
}

func (s *Store) Counts(ctx context.Context, cfg Config) (map[string]int, error) {
	counts := make(map[string]int, len(cfg.Resources))
	for _, name := range cfg.Names() {
		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE resource = ?`, name).Scan(&count); err != nil {
			return nil, fmt.Errorf("count %s: %w", name, err)
		}
		counts[name] = count
	}
	return counts, nil
}

func (s *Store) Seed(ctx context.Context, cfg Config, workspace string, mode SeedMode) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback()
	if mode == SeedReset {
		if _, err := tx.ExecContext(ctx, `DELETE FROM documents`); err != nil {
			return fmt.Errorf("reset business data: %w", err)
		}
	}
	for _, name := range cfg.Names() {
		resource := cfg.Resources[name]
		if resource.Seed == "" {
			continue
		}
		if mode == SeedEmpty {
			var count int
			if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE resource = ?`, name).Scan(&count); err != nil {
				return err
			}
			if count > 0 {
				continue
			}
		}
		documents, err := loadSeed(workspace, resource)
		if err != nil {
			return fmt.Errorf("seed resource %q: %w", name, err)
		}
		for _, document := range documents {
			id := singletonID
			if !resource.Singleton {
				value, ok := document[resource.ID].(string)
				if !ok || value == "" {
					return fmt.Errorf("seed resource %q: every document requires non-empty string field %q", name, resource.ID)
				}
				id = value
			}
			raw, err := json.Marshal(document)
			if err != nil {
				return err
			}
			now := time.Now().UTC().Format(time.RFC3339Nano)
			if _, err := tx.ExecContext(ctx, `INSERT INTO documents(resource, document_id, document_json, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(resource, document_id) DO UPDATE SET document_json = excluded.document_json, updated_at = excluded.updated_at`, name, id, raw, now, now); err != nil {
				return fmt.Errorf("store seed %s/%s: %w", name, id, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}
	return nil
}

func loadSeed(workspace string, resource ResourceDefinition) ([]map[string]any, error) {
	seedPath, err := ResolveSeed(workspace, resource.Seed)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(seedPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if info, err := file.Stat(); err != nil {
		return nil, err
	} else if info.Size() > 1<<20 {
		return nil, errors.New("seed exceeds 1 MiB")
	}
	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.UseNumber()
	if resource.Singleton {
		var document map[string]any
		if err := decoder.Decode(&document); err != nil {
			return nil, fmt.Errorf("singleton seed must be one JSON object: %w", err)
		}
		if err := requireJSONEOF(decoder); err != nil {
			return nil, err
		}
		return []map[string]any{document}, nil
	}
	var documents []map[string]any
	if err := decoder.Decode(&documents); err != nil {
		return nil, fmt.Errorf("collection seed must be a JSON array of objects: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return nil, err
	}
	return documents, nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("seed must contain exactly one JSON value")
		}
		return err
	}
	return nil
}
