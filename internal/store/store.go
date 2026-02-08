package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

const (
	SettingPerPageCents   = "per_page_cents"
	SettingColorPageCents = "color_page_cents"
	SettingRetentionDays  = "retention_days"
)

const DefaultPerPageCents = 10
const DefaultColorPageCents = 30

type Store struct {
	DB *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set journal_mode: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set foreign_keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}

	s := &Store{DB: db}
	if err := s.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) WithTx(ctx context.Context, readOnly bool, fn func(*sql.Tx) error) error {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: readOnly})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			protected INTEGER NOT NULL DEFAULT 0,
			contact_name TEXT,
			phone TEXT,
			email TEXT,
			balance_cents INTEGER NOT NULL DEFAULT 0,
			daily_topup_cents INTEGER NOT NULL DEFAULT 0,
			monthly_topup_cents INTEGER NOT NULL DEFAULT 0,
			yearly_topup_cents INTEGER NOT NULL DEFAULT 0,
			monthly_limit_cents INTEGER NOT NULL DEFAULT 0,
			yearly_limit_cents INTEGER NOT NULL DEFAULT 0,
			month_spent_cents INTEGER NOT NULL DEFAULT 0,
			year_spent_cents INTEGER NOT NULL DEFAULT 0,
			month_period TEXT NOT NULL DEFAULT '',
			year_period TEXT NOT NULL DEFAULT '',
			last_daily_topup TEXT NOT NULL DEFAULT '',
			last_monthly_topup TEXT NOT NULL DEFAULT '',
			last_yearly_topup TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS topups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			amount_cents INTEGER NOT NULL,
			balance_before_cents INTEGER NOT NULL,
			balance_after_cents INTEGER NOT NULL,
			type TEXT NOT NULL,
			operator_user_id INTEGER,
			operator_name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS print_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			printer_uri TEXT NOT NULL,
			filename TEXT NOT NULL,
			stored_path TEXT NOT NULL,
			pages INTEGER NOT NULL,
			cost_cents INTEGER NOT NULL,
			balance_before_cents INTEGER NOT NULL,
			balance_after_cents INTEGER NOT NULL,
			month_total_cents INTEGER NOT NULL,
			year_total_cents INTEGER NOT NULL,
			job_id TEXT,
			status TEXT NOT NULL,
			is_duplex INTEGER NOT NULL DEFAULT 0,
			is_color INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			print_content TEXT,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}

	for _, stmt := range stmts {
		if _, err := s.DB.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	if err := addColumnIfMissing(ctx, s.DB, "users", "protected INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(ctx, s.DB, "print_jobs", "is_duplex INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(ctx, s.DB, "print_jobs", "is_color INTEGER NOT NULL DEFAULT 1"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(ctx, s.DB, "print_jobs", "copies INTEGER NOT NULL DEFAULT 1"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(ctx, s.DB, "print_jobs", "page_range TEXT"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(ctx, s.DB, "print_jobs", "sides TEXT"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(ctx, s.DB, "print_jobs", "duplex TEXT"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if _, err := s.DB.ExecContext(ctx, `INSERT OR IGNORE INTO settings(key, value) VALUES (?, ?), (?, ?), (?, ?)`,
		SettingPerPageCents, strconv.Itoa(DefaultPerPageCents),
		SettingColorPageCents, strconv.Itoa(DefaultColorPageCents),
		SettingRetentionDays, "0",
	); err != nil {
		return fmt.Errorf("seed settings: %w", err)
	}

	return nil
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func addColumnIfMissing(ctx context.Context, db *sql.DB, table string, columnDef string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, columnDef))
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "duplicate column name") {
		return nil
	}
	return err
}
