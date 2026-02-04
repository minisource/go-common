package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Generator helps create new migration files
type Generator struct {
	migrationsDir string
}

// NewGenerator creates a new migration generator
func NewGenerator(migrationsDir string) *Generator {
	return &Generator{migrationsDir: migrationsDir}
}

// Create generates a new migration file pair
func (g *Generator) Create(name string) (string, string, error) {
	// Get next version number
	version, err := g.nextVersion()
	if err != nil {
		return "", "", err
	}

	// Sanitize name
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	name = strings.ReplaceAll(name, "-", "_")

	// Create filenames
	upFile := filepath.Join(g.migrationsDir, fmt.Sprintf("%06d_%s.up.sql", version, name))
	downFile := filepath.Join(g.migrationsDir, fmt.Sprintf("%06d_%s.down.sql", version, name))

	// Create up migration
	if err := g.createFile(upFile, upTemplate, name); err != nil {
		return "", "", err
	}

	// Create down migration
	if err := g.createFile(downFile, downTemplate, name); err != nil {
		os.Remove(upFile) // Cleanup on error
		return "", "", err
	}

	return upFile, downFile, nil
}

// CreateWithTimestamp generates migration with timestamp-based version
func (g *Generator) CreateWithTimestamp(name string) (string, string, error) {
	// Use timestamp as version
	version := time.Now().Format("20060102150405")

	// Sanitize name
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	name = strings.ReplaceAll(name, "-", "_")

	// Create filenames
	upFile := filepath.Join(g.migrationsDir, fmt.Sprintf("%s_%s.up.sql", version, name))
	downFile := filepath.Join(g.migrationsDir, fmt.Sprintf("%s_%s.down.sql", version, name))

	// Create up migration
	if err := g.createFile(upFile, upTemplate, name); err != nil {
		return "", "", err
	}

	// Create down migration
	if err := g.createFile(downFile, downTemplate, name); err != nil {
		os.Remove(upFile) // Cleanup on error
		return "", "", err
	}

	return upFile, downFile, nil
}

func (g *Generator) nextVersion() (int, error) {
	entries, err := os.ReadDir(g.migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(g.migrationsDir, 0755); err != nil {
				return 0, err
			}
			return 1, nil
		}
		return 0, err
	}

	maxVersion := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Extract version number
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion + 1, nil
}

func (g *Generator) createFile(path, tmpl, name string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.Must(template.New("migration").Parse(tmpl))
	return t.Execute(f, map[string]interface{}{
		"Name":      name,
		"Timestamp": time.Now().Format(time.RFC3339),
	})
}

const upTemplate = `-- Migration: {{.Name}}
-- Created at: {{.Timestamp}}
-- Description: Add description here

-- Write your UP migration here
-- Example:
-- CREATE TABLE example (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );
-- CREATE INDEX idx_example_name ON example(name);

`

const downTemplate = `-- Migration: {{.Name}} (ROLLBACK)
-- Created at: {{.Timestamp}}
-- Description: Rollback for {{.Name}}

-- Write your DOWN migration here (reverse of UP)
-- Example:
-- DROP INDEX IF EXISTS idx_example_name;
-- DROP TABLE IF EXISTS example;

`

// ============================================
// Helper for running migrations from GORM DB
// ============================================

// RunFromGormDB runs migrations using underlying sql.DB from GORM
func RunFromGormDB(gormDB interface{ DB() (*sql.DB, error) }, databaseName string) error {
	db, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	migrator, err := NewMigrator(db, databaseName)
	if err != nil {
		return err
	}
	defer migrator.Close()

	return migrator.Up()
}
