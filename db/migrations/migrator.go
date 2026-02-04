package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*.sql
var migrationFiles embed.FS

// Config holds migration configuration
type Config struct {
	DatabaseURL  string
	DatabaseName string
}

// Migrator handles database migrations
type Migrator struct {
	migrate *migrate.Migrate
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB, databaseName string) (*Migrator, error) {
	// Create source driver from embedded files
	sourceDriver, err := iofs.New(migrationFiles, "sql")
	if err != nil {
		return nil, fmt.Errorf("failed to create source driver: %w", err)
	}

	// Create database driver
	dbDriver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: databaseName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, databaseName, dbDriver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

// NewMigratorFromURL creates a migrator from database URL
func NewMigratorFromURL(databaseURL, databaseName string) (*Migrator, error) {
	// Create source driver from embedded files
	sourceDriver, err := iofs.New(migrationFiles, "sql")
	if err != nil {
		return nil, fmt.Errorf("failed to create source driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	err := m.migrate.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Down rolls back all migrations
func (m *Migrator) Down() error {
	err := m.migrate.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	return nil
}

// Steps runs n migrations (positive = up, negative = down)
func (m *Migrator) Steps(n int) error {
	err := m.migrate.Steps(n)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migration steps: %w", err)
	}
	return nil
}

// Version returns current migration version
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Force sets migration version without running migrations
// Use with caution - only for fixing dirty state
func (m *Migrator) Force(version int) error {
	return m.migrate.Force(version)
}

// Close closes the migrator
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}

// MigrationInfo contains migration status information
type MigrationInfo struct {
	Version uint
	Dirty   bool
}

// Status returns current migration status
func (m *Migrator) Status() (*MigrationInfo, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return nil, err
	}
	return &MigrationInfo{
		Version: version,
		Dirty:   dirty,
	}, nil
}
