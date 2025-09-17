package service

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

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

func (s *SQLiteService) Close() error {
	return s.db.Close()
}
