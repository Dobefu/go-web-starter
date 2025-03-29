package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainSuccess(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, main)
}
