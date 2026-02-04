package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var dbConn *sql.DB

// PostgresConfig holds database connection configuration
type PostgresConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DbName          string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// InitDb initializes the database connection
func InitDb(cfg *PostgresConfig) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Tehran",
		cfg.Host, cfg.Port, cfg.User, cfg.Password,
		cfg.DbName, cfg.SSLMode)

	var err error
	dbConn, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	err = dbConn.Ping()
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	// Set connection pool settings
	dbConn.SetMaxIdleConns(cfg.MaxIdleConns)
	dbConn.SetMaxOpenConns(cfg.MaxOpenConns)
	dbConn.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Minute)

	log.Println("Database connection established")
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return dbConn
}

// CloseDB closes the database connection
func CloseDB() error {
	if dbConn != nil {
		return dbConn.Close()
	}
	return nil
}
