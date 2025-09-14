package service

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/luna4dev/airlock/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

const CURRENT_SCHEMA_VERSION = 1

type SQLiteService struct {
	db             *sql.DB
	sqliteSchemaFS *embed.FS
}

func NewSQLiteService(dbPath string, sqliteSchemaFS *embed.FS) (*SQLiteService, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	// Create database file if it doesn't exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file %s: %w", dbPath, err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	service := &SQLiteService{db: db, sqliteSchemaFS: sqliteSchemaFS}

	// Initialize the database schema
	if err := service.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return service, nil
}

func (s *SQLiteService) getSchemaVersion() int {
	var version int
	row := s.db.QueryRow("PRAGMA schema_version")
	row.Scan(&version)
	return version
}

func (s *SQLiteService) initSchema() error {
	currentVersion := s.getSchemaVersion()

	if currentVersion < CURRENT_SCHEMA_VERSION {
		// Read schema from external file
		schemaFileName := fmt.Sprintf("configs/sqlite-schema/schema-v%d.sql", CURRENT_SCHEMA_VERSION)
		schemaBytes, err := s.sqliteSchemaFS.ReadFile(schemaFileName)
		if err != nil {
			return fmt.Errorf("failed to read schema file %s: %w", schemaFileName, err)
		}

		// apply schema
		schema := string(schemaBytes)
		_, err = s.db.Exec(schema)
		if err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

func (s *SQLiteService) GetAllUsers(ctx context.Context) ([]*model.Luna4User, error) {
	query := `
		SELECT id, email, status, created_at, updated_at, last_login_at
		FROM luna4_users
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*model.Luna4User
	for rows.Next() {
		var user model.Luna4User
		var lastLoginAt sql.NullInt64

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Int64
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return users, nil
}

func (s *SQLiteService) CreateUser(ctx context.Context, user *model.Luna4User) error {
	query := `
		INSERT INTO luna4_users (id, email, status, created_at, updated_at, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
	)

	return err
}

func (s *SQLiteService) GetUserByID(ctx context.Context, userID string) (*model.Luna4User, error) {
	query := `
		SELECT id, email, status, created_at, updated_at, last_login_at
		FROM luna4_users
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, userID)

	var user model.Luna4User
	var lastLoginAt sql.NullInt64

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Int64
	}

	return &user, nil
}

func (s *SQLiteService) UpdateUserStatus(ctx context.Context, userID string, status model.UserStatus) error {
	query := `
		UPDATE luna4_users
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx, query, status, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

func (s *SQLiteService) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM luna4_users WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID: %s", userID)
	}

	return nil
}

func (s *SQLiteService) Close() error {
	return s.db.Close()
}
