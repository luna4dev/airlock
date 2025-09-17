package service

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const CURRENT_SCHEMA_VERSION = 2

type SQLiteService struct {
	db                *sql.DB
	sqliteSchemaFS    *embed.FS
	sqliteMigrationFS *embed.FS
}

func NewSQLiteService(dbPath string, sqliteSchemaFS *embed.FS, sqliteMigrationFS *embed.FS) (*SQLiteService, error) {
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

	service := &SQLiteService{db: db, sqliteSchemaFS: sqliteSchemaFS, sqliteMigrationFS: sqliteMigrationFS}

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
	// Note that schema version is set at the end of the schema-v#.sql file
	currentVersion := s.getSchemaVersion()

	if currentVersion < CURRENT_SCHEMA_VERSION {
		log.Printf("Database migration required: current version %d, target version %d", currentVersion, CURRENT_SCHEMA_VERSION)
		// Recursively migrate up to current version
		if err := s.migrateToVersion(currentVersion + 1); err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
		log.Printf("Database migration completed successfully to version %d", CURRENT_SCHEMA_VERSION)
	} else {
		log.Printf("Database schema is up to date (version %d)", currentVersion)
	}

	// read schema and sync
	schemaFileName := fmt.Sprintf("configs/sqlite-schema/schema-v%d.sql", CURRENT_SCHEMA_VERSION)
	schemaBytes, err := s.sqliteSchemaFS.ReadFile(schemaFileName)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", schemaFileName, err)
	}

	// apply schema
	schema := string(schemaBytes)
	_, err = s.db.Exec(schema)
	if err != nil {
		// TODO: if the flow fails here because of the SQL syntax error
		//		 the schema_version bump up (should not happen). Prevent the version bump
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

func (s *SQLiteService) migrateToVersion(targetVersion int) error {
	if targetVersion > CURRENT_SCHEMA_VERSION {
		return nil
	}

	log.Printf("Applying database migration to version %d", targetVersion)

	// Apply migration for target version
	migrationFileName := fmt.Sprintf("configs/sqlite-migration/migration-v%d.sql", targetVersion)
	migrationBytes, err := s.sqliteMigrationFS.ReadFile(migrationFileName)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", migrationFileName, err)
	}

	// Execute migration
	migration := string(migrationBytes)
	_, err = s.db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to execute migration v%d: %w", targetVersion, err)
	}

	log.Printf("Successfully applied migration to version %d", targetVersion)

	// Recursively migrate to next version if needed
	if targetVersion < CURRENT_SCHEMA_VERSION {
		return s.migrateToVersion(targetVersion + 1)
	}

	return nil
}

func (s *SQLiteService) Close() error {
	return s.db.Close()
}
