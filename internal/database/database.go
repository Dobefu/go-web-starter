package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	_ "github.com/lib/pq"
)

type DatabaseInterface interface {
	Close() error
	Ping() error
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Begin() (*sql.Tx, error)
	Stats() sql.DBStats
}

type Database struct {
	db     DatabaseInterface
	logger *logger.Logger
}

var errNotInitialized error = fmt.Errorf("database not initialized")

var New = func(cfg config.Database, log *logger.Logger) (*Database, error) {
	log.Debug("Initializing database connection", logger.Fields{
		"host":   cfg.Host,
		"port":   cfg.Port,
		"dbname": cfg.DBName,
	})

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	log.Trace("Testing database connection", nil)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Debug("Database connection established", nil)

	return &Database{
		db:     db,
		logger: log,
	}, nil
}

func (d *Database) Close() error {
	if d.db == nil {
		return errNotInitialized
	}

	if d.logger != nil {
		d.logger.Debug("Closing database connection", nil)
	}

	return d.db.Close()
}

func (d *Database) Ping() error {
	if d.db == nil {
		return errNotInitialized
	}

	return d.db.Ping()
}

func (d *Database) Query(query string, args ...any) (*sql.Rows, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	if d.logger != nil {
		d.logger.Debug("Executing database query", logger.Fields{
			"query": query,
			"args":  args,
		})
	}

	rows, err := d.db.Query(query, args...)

	if err != nil && d.logger != nil {
		d.logger.Error("Database query failed", logger.Fields{
			"query": query,
			"error": err.Error(),
		})
	}

	return rows, err
}

func (d *Database) QueryRow(query string, args ...any) *sql.Row {
	if d.db == nil {
		return nil
	}

	if d.logger != nil {
		d.logger.Debug("Executing database query row", logger.Fields{
			"query": query,
			"args":  args,
		})
	}

	return d.db.QueryRow(query, args...)
}

func (d *Database) Exec(query string, args ...any) (sql.Result, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	if d.logger != nil {
		d.logger.Debug("Executing database command", logger.Fields{
			"query": query,
			"args":  args,
		})
	}

	result, err := d.db.Exec(query, args...)

	if err != nil && d.logger != nil {
		d.logger.Error("Database command failed", logger.Fields{
			"query": query,
			"error": err.Error(),
		})
	}

	return result, err
}

func (d *Database) Begin() (*sql.Tx, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.Begin()
}

func (d *Database) Stats() sql.DBStats {
	if d.db == nil {
		return sql.DBStats{}
	}

	return d.db.Stats()
}
