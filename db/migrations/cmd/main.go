package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

/*
Migration CLI Tool
==================

This is a standalone CLI tool for managing database migrations.
Each service should copy this file to their cmd/migrate/main.go

Usage:
  go run cmd/migrate/main.go -command=up
  go run cmd/migrate/main.go -command=down
  go run cmd/migrate/main.go -command=down -steps=1
  go run cmd/migrate/main.go -command=version
  go run cmd/migrate/main.go -command=force -version=1
  go run cmd/migrate/main.go -command=create -name=add_users_table

Environment Variables:
  DATABASE_URL - PostgreSQL connection string
  MIGRATIONS_PATH - Path to migrations folder (default: ./migrations)
*/

func main() {
	// Parse flags
	command := flag.String("command", "", "Migration command: up, down, version, force, create, status")
	steps := flag.Int("steps", 0, "Number of steps for step-based migration")
	version := flag.Int("version", 0, "Version number for force command")
	name := flag.String("name", "", "Migration name for create command")
	migrationsPath := flag.String("path", "", "Path to migrations folder")
	databaseURL := flag.String("database", "", "Database URL (overrides DATABASE_URL env)")
	flag.Parse()

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	// Get database URL
	dbURL := *databaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}

	// Get migrations path
	migPath := *migrationsPath
	if migPath == "" {
		migPath = os.Getenv("MIGRATIONS_PATH")
	}
	if migPath == "" {
		migPath = "./migrations"
	}

	switch *command {
	case "create":
		if *name == "" {
			log.Fatal("Error: -name flag is required for create command")
		}
		createMigration(migPath, *name)

	case "up", "down", "version", "force", "status":
		if dbURL == "" {
			log.Fatal("Error: DATABASE_URL environment variable or -database flag is required")
		}
		runMigrationCommand(*command, dbURL, migPath, *steps, *version)

	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func runMigrationCommand(command, dbURL, migPath string, steps, forceVersion int) {
	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create postgres driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create database driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migPath),
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	switch command {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		printVersion(m)
		log.Println("✓ Migrations applied successfully")

	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		printVersion(m)
		log.Println("✓ Migrations rolled back successfully")

	case "version":
		printVersion(m)

	case "status":
		printStatus(m)

	case "force":
		if forceVersion == 0 {
			log.Fatal("Error: -version flag is required for force command")
		}
		if err := m.Force(forceVersion); err != nil {
			log.Fatalf("Force version failed: %v", err)
		}
		log.Printf("✓ Forced version to %d", forceVersion)
	}
}

func printVersion(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			log.Println("No migrations applied yet")
			return
		}
		log.Fatalf("Failed to get version: %v", err)
	}
	dirtyStr := ""
	if dirty {
		dirtyStr = " (DIRTY - migration failed, needs manual fix)"
	}
	log.Printf("Current version: %d%s", version, dirtyStr)
}

func printStatus(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("╔════════════════════════════════════════╗")
			fmt.Println("║       Migration Status: CLEAN          ║")
			fmt.Println("╠════════════════════════════════════════╣")
			fmt.Println("║ No migrations applied yet              ║")
			fmt.Println("║ Run 'migrate -command=up' to apply     ║")
			fmt.Println("╚════════════════════════════════════════╝")
			return
		}
		log.Fatalf("Failed to get status: %v", err)
	}

	status := "CLEAN"
	if dirty {
		status = "DIRTY"
	}

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Printf("║       Migration Status: %-14s ║\n", status)
	fmt.Println("╠════════════════════════════════════════╣")
	fmt.Printf("║ Current Version: %-21d ║\n", version)
	if dirty {
		fmt.Println("║ ⚠ Database is in dirty state!          ║")
		fmt.Println("║ Run 'migrate -command=force -version=N'║")
		fmt.Println("║ to fix manually                        ║")
	}
	fmt.Println("╚════════════════════════════════════════╝")
}

func createMigration(path, name string) {
	// Ensure migrations directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Get next version
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	version := 1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var v int
		if _, err := fmt.Sscanf(entry.Name(), "%d_", &v); err == nil {
			if v >= version {
				version = v + 1
			}
		}
	}

	// Create migration files
	upFile := fmt.Sprintf("%s/%06d_%s.up.sql", path, version, name)
	downFile := fmt.Sprintf("%s/%06d_%s.down.sql", path, version, name)

	upContent := fmt.Sprintf(`-- Migration: %s
-- Version: %d

-- Write your UP migration here

`, name, version)

	downContent := fmt.Sprintf(`-- Rollback: %s
-- Version: %d

-- Write your DOWN migration here (reverse of UP)

`, name, version)

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatalf("Failed to create up migration: %v", err)
	}

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		os.Remove(upFile)
		log.Fatalf("Failed to create down migration: %v", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Migration Files Created Successfully             ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ UP:   %s\n", upFile)
	fmt.Printf("║ DOWN: %s\n", downFile)
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

func printUsage() {
	fmt.Println(`
╔══════════════════════════════════════════════════════════════════════╗
║                     Database Migration Tool                          ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  COMMANDS:                                                           ║
║    up        Apply all pending migrations                            ║
║    down      Rollback all migrations                                 ║
║    version   Show current migration version                          ║
║    status    Show detailed migration status                          ║
║    force     Force set version (fix dirty state)                     ║
║    create    Create new migration files                              ║
║                                                                      ║
║  FLAGS:                                                              ║
║    -command   Migration command (required)                           ║
║    -steps     Number of steps for up/down (optional)                 ║
║    -version   Version number for force command                       ║
║    -name      Migration name for create command                      ║
║    -path      Path to migrations folder (default: ./migrations)      ║
║    -database  Database URL (overrides DATABASE_URL env)              ║
║                                                                      ║
║  EXAMPLES:                                                           ║
║    migrate -command=create -name=add_users_table                     ║
║    migrate -command=up                                               ║
║    migrate -command=down -steps=1                                    ║
║    migrate -command=status                                           ║
║    migrate -command=force -version=5                                 ║
║                                                                      ║
║  ENVIRONMENT:                                                        ║
║    DATABASE_URL      PostgreSQL connection string                    ║
║    MIGRATIONS_PATH   Path to migrations folder                       ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝`)
}
