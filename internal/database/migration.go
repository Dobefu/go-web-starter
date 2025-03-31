package database

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/Dobefu/go-web-starter/internal/logger"
)

type FS interface {
	fs.FS
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
}

//go:embed migrations/*
var content embed.FS

type ContentFS struct {
	content FS
}

var contentFS = ContentFS{content: content}

func MigrateDown(cfg Config) (err error) {
	logInfo := logger.New(logger.InfoLevel, os.Stdout)
	dbConn, _ := New(cfg, nil)

	version, _ := getMigrationState(dbConn)

	if version == 0 {
		logInfo.Info("Nothing to revert", nil)
		return nil
	}

	files, err := contentFS.content.ReadDir("migrations")

	if err != nil {
		return err
	}

	migrationIndex := len(files) / 2

	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		name := file.Name()

		if strings.Split(name, ".")[1] != "down" {
			continue
		}

		migrationIndex = migrationIndex - 1

		if migrationIndex >= version {
			continue
		}

		logInfo.Info(fmt.Sprintf("Running migration: %s", name), nil)
		err = runMigration(dbConn, name, migrationIndex)

		if err != nil {
			return err
		}
	}

	err = setMigrationState(dbConn, migrationIndex, false)

	if err != nil {
		return err
	}

	return nil
}

func MigrateUp(cfg Config) (err error) {
	logInfo := logger.New(logger.InfoLevel, os.Stdout)
	dbConn, _ := New(cfg, nil)

	err = createMigrationsTable(dbConn)

	if err != nil {
		return err
	}

	version, _ := getMigrationState(dbConn)

	files, err := contentFS.content.ReadDir("migrations")

	if err != nil {
		return err
	}

	migrationIndex := 0

	for _, file := range files {
		name := file.Name()

		if strings.Split(name, ".")[1] != "up" {
			continue
		}

		migrationIndex += 1

		if migrationIndex <= version {
			continue
		}

		logInfo.Info(fmt.Sprintf("Running migration: %s", name), nil)
		err = runMigration(dbConn, name, migrationIndex)

		if err != nil {
			return err
		}
	}

	err = setMigrationState(dbConn, migrationIndex, false)

	if err != nil {
		return err
	}

	return nil
}

func MigrateVersion(cfg Config) (version int, err error) {
	dbConn, _ := New(cfg, nil)

	row, err := dbConn.QueryRow("SELECT version FROM migrations LIMIT 1")

	if err != nil {
		return 0, err
	}

	err = row.Scan(&version)

	if err != nil {
		return 0, err
	}

	return version, nil
}

func createMigrationsTable(dbConn *Database) (err error) {
	_, err = dbConn.Exec(`
    CREATE TABLE IF NOT EXISTS migrations(
      version bigint NOT NULL PRIMARY KEY,
      dirty boolean NOT NULL
    );
  `)

	if err != nil {
		return err
	}

	return nil
}

func getMigrationState(dbConn *Database) (version int, dirty bool) {
	row, err := dbConn.QueryRow("SELECT version,dirty FROM migrations LIMIT 1")

	if err != nil {
		return 0, true
	}

	err = row.Scan(&version, &dirty)

	// If nothing is found, the table is empty.
	// This is fine, since an initial migration will produce this result.
	// When this happens, default values should be returned.
	if err != nil {
		return 0, false
	}

	return version, dirty
}

func setMigrationState(dbConn *Database, version int, dirty bool) (err error) {
	_, err = dbConn.Exec("TRUNCATE migrations")

	if err != nil {
		return err
	}

	_, err = dbConn.Exec("INSERT INTO migrations (version, dirty) VALUES ($1, $2)", version, dirty)

	if err != nil {
		return err
	}

	return nil
}

func runMigration(dbConn *Database, filename string, index int) (err error) {
	queryBytes, err := contentFS.content.ReadFile(fmt.Sprintf("migrations/%s", filename))

	if err != nil {
		_ = setMigrationState(dbConn, index, true)
		return err
	}

	_, err = dbConn.Exec(string(queryBytes))

	if err != nil {
		_ = setMigrationState(dbConn, index, true)
		return err
	}

	return nil
}
