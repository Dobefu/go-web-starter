package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainSuccess(t *testing.T) {
	assert.NotPanics(t, main)
}
