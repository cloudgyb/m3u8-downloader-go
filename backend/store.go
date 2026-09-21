package backend

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps the local SQLite database used for settings and history.
type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) the database in the OS config directory.
func NewStore() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	dir = filepath.Join(dir, "m3u8-downloader")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "m3u8dl.db"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: serialize access

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS history (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL,
			url         TEXT NOT NULL,
			size        INTEGER NOT NULL DEFAULT 0,
			status      TEXT NOT NULL,
			finished_at TEXT NOT NULL,
			duration    TEXT NOT NULL DEFAULT '',
			error       TEXT NOT NULL DEFAULT '',
			file        TEXT NOT NULL DEFAULT ''
		);
	`)
	if err != nil {
		return err
	}
	// Upgrade pre-existing databases that predate the `file` column.
	if !s.hasColumn("history", "file") {
		_, _ = s.db.Exec(`ALTER TABLE history ADD COLUMN file TEXT NOT NULL DEFAULT ''`)
	}
	return nil
}

func (s *Store) hasColumn(table, column string) bool {
	rows, err := s.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

func (s *Store) Close() error { return s.db.Close() }

// ── settings ────────────────────────────────────────────────────────────

func (s *Store) GetSettings() (Settings, error) {
	cfg := defaultSettings()
	var raw string
	err := s.db.QueryRow(`SELECT value FROM kv WHERE key = 'settings'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, err
	}
	if cfg.SaveDir == "" {
		cfg.SaveDir = defaultSettings().SaveDir
	}
	if cfg.Threads <= 0 {
		cfg.Threads = 8
	}
	return cfg, nil
}

func (s *Store) SaveSettings(cfg Settings) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO kv (key, value) VALUES ('settings', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		string(b),
	)
	return err
}

// ── history ─────────────────────────────────────────────────────────────

func (s *Store) GetHistory() ([]HistoryItem, error) {
	rows, err := s.db.Query(
		`SELECT id, name, url, size, status, finished_at, duration, error, file
		 FROM history ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]HistoryItem, 0)
	for rows.Next() {
		var h HistoryItem
		if err := rows.Scan(&h.ID, &h.Name, &h.URL, &h.Size, &h.Status, &h.FinishedAt, &h.Duration, &h.Error, &h.File); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

func (s *Store) AddHistory(h HistoryItem) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO history (name, url, size, status, finished_at, duration, error, file)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		h.Name, h.URL, h.Size, h.Status, h.FinishedAt, h.Duration, h.Error, h.File,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeleteHistory(id int64) (int64, error) {
	exec, err := s.db.Exec(`DELETE FROM history WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	affected, err := exec.RowsAffected()
	if err != nil {
		return affected, err
	}
	return affected, nil
}

func (s *Store) ClearHistory() error {
	_, err := s.db.Exec(`DELETE FROM history`)
	return err
}
