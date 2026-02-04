package gormdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormConfig holds database configuration
type GormConfig struct {
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

// Database wraps the gorm.DB client
type Database struct {
	db  *gorm.DB
	cfg *GormConfig
}

// NewDatabase creates and returns a new Database instance
func NewDatabase(cfg *GormConfig) (*Database, error) {
	cnn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Tehran",
		cfg.Host, cfg.Port, cfg.User, cfg.Password,
		cfg.DbName, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(cnn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	if err := sqlDb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDb.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Minute)

	log.Println("Db connection established")
	return &Database{db: db, cfg: cfg}, nil
}

// DB returns the underlying gorm.DB instance
func (d *Database) DB() *gorm.DB {
	return d.db
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDb, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDb.Close()
}

// Ping checks database connectivity
func (d *Database) Ping(ctx context.Context) error {
	sqlDb, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDb.PingContext(ctx)
}

// WithContext returns a new DB with context
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// Deprecated: Use NewDatabase instead
// The following functions are kept for backward compatibility

var dbClient *gorm.DB

// InitDb initializes the database connection (deprecated: use NewDatabase)
func InitDb(cfg *GormConfig) error {
	var err error
	cnn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Tehran",
		cfg.Host, cfg.Port, cfg.User, cfg.Password,
		cfg.DbName, cfg.SSLMode)

	dbClient, err = gorm.Open(postgres.Open(cnn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDb, _ := dbClient.DB()
	err = sqlDb.Ping()
	if err != nil {
		return err
	}

	sqlDb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDb.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Minute)

	log.Println("Db connection established")
	return nil
}

// GetDb returns the database client (deprecated: use NewDatabase)
func GetDb() *gorm.DB {
	return dbClient
}

// CloseDb closes the database connection (deprecated: use Database.Close)
func CloseDb() {
	con, _ := dbClient.DB()
	con.Close()
}
