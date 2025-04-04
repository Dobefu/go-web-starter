package config

import (
	"testing"

	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetLogLevelOverride(t *testing.T) {
	viper.Reset()

	viper.Set("log.level", 1)

	logLevel := GetLogLevel()
	assert.Equal(t, logger.DebugLevel, logLevel)
}

func TestGetLogLevel(t *testing.T) {
	viper.Reset()

	defaultLogLevel := GetLogLevel()
	assert.Equal(t, logger.Level(DefaultConfig.Log.Level), defaultLogLevel)
}
