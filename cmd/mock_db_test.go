package cmd

import (
	"database/sql"
)

type mockDB struct{}

func (m *mockDB) Close() error                                       { return nil }
func (m *mockDB) Ping() error                                        { return nil }
func (m *mockDB) Query(query string, args ...any) (*sql.Rows, error) { return nil, nil }
func (m *mockDB) QueryRow(query string, args ...any) *sql.Row        { return (*sql.Row)(nil) }
func (m *mockDB) Exec(query string, args ...any) (sql.Result, error) { return nil, nil }
func (m *mockDB) Begin() (*sql.Tx, error)                            { return nil, nil }
func (m *mockDB) Stats() sql.DBStats                                 { return sql.DBStats{} }
