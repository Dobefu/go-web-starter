package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfigSuccess(t *testing.T) {
	assert.NotPanics(t, initConfig)
}

func TestExecuteSuccess(t *testing.T) {
	assert.NotPanics(t, Execute)
}
