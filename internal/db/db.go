package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// DB is a wrapper around sqlx.DB that provides additional functionality
type DB struct {
	*sqlx.DB
}

// Status represents a user status in the database
type Status struct {
	URI       string `db:"uri"`
	AuthorDID string `db:"authorDid"`
	Status    string `db:"status"`
	CreatedAt string `db:"createdAt"`
	IndexedAt string `db:"indexedAt"`
}

// AuthSession represents an authentication session in the database
type AuthSession struct {
	Key     string `db:"key"`
	Session string `db:"session"`
}

// AuthState represents an authentication state in the database
type AuthState struct {
	Key   string `db:"key"`
	State string `db:"state"`
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	// Connect to the SQLite database
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure the connection
	db.SetMaxOpenConns(1) // SQLite supports only one writer at a time

	return &DB{DB: db}, nil
}

// Migrate runs database migrations
func (db *DB) Migrate() error {
	log.Info().Msg("Running database migrations...")

	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS status (
		uri TEXT PRIMARY KEY,
		authorDid TEXT NOT NULL,
		status TEXT NOT NULL,
		createdAt TEXT NOT NULL,
		indexedAt TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS auth_session (
		key TEXT PRIMARY KEY,
		session TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS auth_state (
		key TEXT PRIMARY KEY,
		state TEXT NOT NULL
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Info().Msg("Database migrations completed successfully")
	return nil
}

// GetRecentStatuses retrieves recent statuses from the database
func (db *DB) GetRecentStatuses(limit int) ([]Status, error) {
	var statuses []Status

	query := `
	SELECT * FROM status
	ORDER BY indexedAt DESC
	LIMIT ?
	`

	err := db.Select(&statuses, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent statuses: %w", err)
	}

	return statuses, nil
}

// GetUserStatus retrieves the latest status for a user
func (db *DB) GetUserStatus(authorDID string) (*Status, error) {
	var status Status

	query := `
	SELECT * FROM status
	WHERE authorDid = ?
	ORDER BY indexedAt DESC
	LIMIT 1
	`

	err := db.Get(&status, query, authorDID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user status: %w", err)
	}

	return &status, nil
}

// SaveStatus stores a status in the database
func (db *DB) SaveStatus(status *Status) error {
	query := `
	INSERT INTO status (uri, authorDid, status, createdAt, indexedAt)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT (uri) DO UPDATE SET
		status = excluded.status,
		indexedAt = excluded.indexedAt
	`

	_, err := db.Exec(
		query,
		status.URI,
		status.AuthorDID,
		status.Status,
		status.CreatedAt,
		status.IndexedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// DeleteStatus removes a status from the database
func (db *DB) DeleteStatus(uri string) error {
	query := `DELETE FROM status WHERE uri = ?`

	_, err := db.Exec(query, uri)
	if err != nil {
		return fmt.Errorf("failed to delete status: %w", err)
	}

	return nil
}

// The following methods are for auth session storage

// GetAuthSession retrieves an auth session from the database
func (db *DB) GetAuthSession(key string) (string, error) {
	var session AuthSession

	query := `SELECT * FROM auth_session WHERE key = ?`

	err := db.Get(&session, query, key)
	if err != nil {
		return "", fmt.Errorf("failed to get auth session: %w", err)
	}

	return session.Session, nil
}

// SaveAuthSession stores an auth session in the database
func (db *DB) SaveAuthSession(key, sessionData string) error {
	query := `
	INSERT INTO auth_session (key, session)
	VALUES (?, ?)
	ON CONFLICT (key) DO UPDATE SET
		session = excluded.session
	`

	_, err := db.Exec(query, key, sessionData)
	if err != nil {
		return fmt.Errorf("failed to save auth session: %w", err)
	}

	return nil
}

// DeleteAuthSession removes an auth session from the database
func (db *DB) DeleteAuthSession(key string) error {
	query := `DELETE FROM auth_session WHERE key = ?`

	_, err := db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete auth session: %w", err)
	}

	return nil
}

// The following methods are for auth state storage

// GetAuthState retrieves an auth state from the database
func (db *DB) GetAuthState(key string) (string, error) {
	var state AuthState

	query := `SELECT * FROM auth_state WHERE key = ?`

	err := db.Get(&state, query, key)
	if err != nil {
		return "", fmt.Errorf("failed to get auth state: %w", err)
	}

	return state.State, nil
}

// SaveAuthState stores an auth state in the database
func (db *DB) SaveAuthState(key, stateData string) error {
	query := `
	INSERT INTO auth_state (key, state)
	VALUES (?, ?)
	ON CONFLICT (key) DO UPDATE SET
		state = excluded.state
	`

	_, err := db.Exec(query, key, stateData)
	if err != nil {
		return fmt.Errorf("failed to save auth state: %w", err)
	}

	return nil
}

// DeleteAuthState removes an auth state from the database
func (db *DB) DeleteAuthState(key string) error {
	query := `DELETE FROM auth_state WHERE key = ?`

	_, err := db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete auth state: %w", err)
	}

	return nil
}