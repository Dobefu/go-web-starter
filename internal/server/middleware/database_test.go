package middleware

import (
	"database/sql"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockDatabase struct {
	database.DatabaseInterface
}

func (m *MockDatabase) Close() error                                       { return nil }
func (m *MockDatabase) Ping() error                                        { return nil }
func (m *MockDatabase) Query(query string, args ...any) (*sql.Rows, error) { return nil, nil }
func (m *MockDatabase) QueryRow(query string, args ...any) *sql.Row        { return nil }
func (m *MockDatabase) Exec(query string, args ...any) (sql.Result, error) { return nil, nil }
func (m *MockDatabase) Begin() (*sql.Tx, error)                            { return nil, nil }

func TestDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := &MockDatabase{}

	c, _ := gin.CreateTestContext(nil)

	middleware := Database(mockDB)
	middleware(c)

	db, exists := c.Get("db")
	assert.True(t, exists)
	assert.Equal(t, mockDB, db)
}
