package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloseNotInitialized(t *testing.T) {
	database := Database{}

	err := database.Close()
	assert.EqualError(t, err, errNotInitialized.Error())
}

func TestPingNotInitialized(t *testing.T) {
	database := Database{}

	err := database.Ping()
	assert.EqualError(t, err, errNotInitialized.Error())
}

func TestQueryNotInitialized(t *testing.T) {
	database := Database{}

	rows, err := database.Query("", nil)
	assert.EqualError(t, err, errNotInitialized.Error())
	assert.Nil(t, rows)
}

func TestQueryRowNotInitialized(t *testing.T) {
	database := Database{}

	row, err := database.QueryRow("", nil)
	assert.EqualError(t, err, errNotInitialized.Error())
	assert.Nil(t, row)
}

func TestExecNotInitialized(t *testing.T) {
	database := Database{}

	result, err := database.Exec("", nil)
	assert.EqualError(t, err, errNotInitialized.Error())
	assert.Nil(t, result)
}

func TestBeginNotInitialized(t *testing.T) {
	database := Database{}

	tx, err := database.Begin()
	assert.EqualError(t, err, errNotInitialized.Error())
	assert.Nil(t, tx)
}
