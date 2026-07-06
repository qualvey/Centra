package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"eventguard/internal/core"

	_ "modernc.org/sqlite"
)

type IPState struct {
	IP        string
	Score     int
	Count     int
	LastSeen  time.Time
	Status    string
	UpdatedAt time.Time
}

type SQLiteStore struct {
	db *sql.DB
}

func OpenSQLite(path string) (*SQLiteStore, error) {
	if path == "" {
		path = ".eventguard/eventguard.db"
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create database dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStore) SaveEvent(ctx context.Context, event core.Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("marshal event metadata: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO events (
			timestamp, source, service, event_type, level, ip, message, metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		formatTime(event.Timestamp),
		event.Source,
		event.Service,
		event.EventType,
		event.Level,
		event.IP,
		event.Message,
		string(metadata),
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	if event.IP != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO ip_states (
				ip, score, count, last_seen, status, updated_at
			) VALUES (?, 0, 1, ?, 'observed', ?)
			ON CONFLICT(ip) DO UPDATE SET
				count = count + 1,
				last_seen = excluded.last_seen,
				updated_at = excluded.updated_at
		`, event.IP, formatTime(event.Timestamp), formatTime(time.Now().UTC()))
		if err != nil {
			return fmt.Errorf("upsert ip state: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit event: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Increment(ctx context.Context, key string) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO counters (key, value) VALUES (?, 1)
		ON CONFLICT(key) DO UPDATE SET value = value + 1
	`, key)
	if err != nil {
		return 0, fmt.Errorf("increment counter: %w", err)
	}

	var value int
	if err := tx.QueryRowContext(ctx, `SELECT value FROM counters WHERE key = ?`, key).Scan(&value); err != nil {
		return 0, fmt.Errorf("read counter: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit counter: %w", err)
	}
	return value, nil
}

func (s *SQLiteStore) MarkOnce(ctx context.Context, key string) (bool, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO marks (key, created_at) VALUES (?, ?)
	`, key, formatTime(time.Now().UTC()))
	if err != nil {
		return false, fmt.Errorf("insert mark: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read mark result: %w", err)
	}
	return affected == 1, nil
}

func (s *SQLiteStore) CountEvents(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *SQLiteStore) GetIPState(ctx context.Context, ip string) (IPState, bool, error) {
	var state IPState
	var lastSeen, updatedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT ip, score, count, last_seen, status, updated_at
		FROM ip_states
		WHERE ip = ?
	`, ip).Scan(&state.IP, &state.Score, &state.Count, &lastSeen, &state.Status, &updatedAt)
	if err == sql.ErrNoRows {
		return IPState{}, false, nil
	}
	if err != nil {
		return IPState{}, false, err
	}

	parsedLastSeen, err := time.Parse(time.RFC3339Nano, lastSeen)
	if err != nil {
		return IPState{}, false, fmt.Errorf("parse last_seen: %w", err)
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return IPState{}, false, fmt.Errorf("parse updated_at: %w", err)
	}
	state.LastSeen = parsedLastSeen
	state.UpdatedAt = parsedUpdatedAt
	return state, true, nil
}

func (s *SQLiteStore) migrate(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			source TEXT NOT NULL,
			service TEXT NOT NULL,
			event_type TEXT NOT NULL,
			level TEXT NOT NULL,
			ip TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata_json TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_events_ip ON events(ip)`,
		`CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type)`,
		`CREATE TABLE IF NOT EXISTS ip_states (
			ip TEXT PRIMARY KEY,
			score INTEGER NOT NULL DEFAULT 0,
			count INTEGER NOT NULL DEFAULT 0,
			last_seen TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'observed',
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS counters (
			key TEXT PRIMARY KEY,
			value INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS marks (
			key TEXT PRIMARY KEY,
			created_at TEXT NOT NULL
		)`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate sqlite database: %w", err)
		}
	}
	return nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
