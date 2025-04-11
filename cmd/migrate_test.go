package cmd

import (
	"testing"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetDatabaseConfig(t *testing.T) {
	cfg := getDatabaseConfig()
	assert.Equal(t, config.Database{}, cfg)
}
