package cmd

import (
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
)

func TestGetNoColorFromEnv(t *testing.T) {
	assert.False(t, getNoColorFromEnv())

	t.Setenv("NO_COLOR", "")
	assert.False(t, getNoColorFromEnv())

	t.Setenv("NO_COLOR", "1")
	assert.True(t, getNoColorFromEnv())

	t.Setenv("NO_COLOR", "foo")
	assert.True(t, getNoColorFromEnv())
}
