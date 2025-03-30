package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Dobefu/go-web-starter/internal/logger"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type Database struct {
	db     *sql.DB
	logger *logger.Logger
}

func New(cfg Config, log *logger.Logger) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{
		db:     db,
		logger: log,
	}, nil
}

func (d *Database) Close() error {
	if d.db == nil {
		return nil
	}

	return d.db.Close()
}

func (d *Database) Ping() error {
	if d.db == nil {
		return fmt.Errorf("database not initialized")
	}

	return d.db.Ping()
}

func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	return d.db.Query(query, args...)
}

func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	if d.db == nil {
		return nil
	}

	return d.db.QueryRow(query, args...)
}

func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	return d.db.Exec(query, args...)
}

func (d *Database) Begin() (*sql.Tx, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	return d.db.Begin()
}
