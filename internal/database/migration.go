package database

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/Dobefu/go-web-starter/internal/config"
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

const errFmtFailedToInitDB = "failed to initialize database connection: %v"

func MigrateDown(cfg config.Database) (err error) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	dbConn, err := New(cfg, nil)

	if err != nil || dbConn == nil {
		return fmt.Errorf(errFmtFailedToInitDB, err)
	}

	db, _ := dbConn.(*Database)

	version, dirty := getMigrationState(db)

	if dirty {
		return fmt.Errorf("the migrations table is in a dirty state")
	}

	if version == 0 {
		log.Info("Nothing to revert", nil)
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

		log.Info(fmt.Sprintf("Running migration: %s", name), nil)
		err = runMigration(db, name, migrationIndex)

		if err != nil {
			return err
		}
	}

	err = setMigrationState(db, migrationIndex, false)

	if err != nil {
		return err
	}

	return nil
}

func MigrateUp(cfg config.Database) (err error) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	dbConn, err := New(cfg, nil)

	if err != nil || dbConn == nil {
		return fmt.Errorf(errFmtFailedToInitDB, err)
	}

	db, _ := dbConn.(*Database)

	err = createMigrationsTable(db)

	if err != nil {
		return err
	}

	version, dirty := getMigrationState(db)

	if dirty {
		return fmt.Errorf("the migrations table is in a dirty state")
	}

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

		log.Info(fmt.Sprintf("Running migration: %s", name), nil)
		err = runMigration(db, name, migrationIndex)

		if err != nil {
			return err
		}
	}

	err = setMigrationState(db, migrationIndex, false)

	if err != nil {
		return err
	}

	return nil
}

func MigrateVersion(cfg config.Database) (version int, err error) {
	dbConn, err := New(cfg, nil)

	if err != nil || dbConn == nil {
		return 0, fmt.Errorf(errFmtFailedToInitDB, err)
	}

	row := dbConn.QueryRow("SELECT version FROM migrations LIMIT 1")
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
	row := dbConn.QueryRow("SELECT version,dirty FROM migrations LIMIT 1")
	err := row.Scan(&version, &dirty)

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
