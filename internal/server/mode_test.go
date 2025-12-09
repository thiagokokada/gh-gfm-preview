package server

import (
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
)

func TestGetMode(t *testing.T) {
	param := &Param{ForceLightMode: true}
	modeString := param.getMode().String()
	expected := "light"

	assert.Equal(t, modeString, expected)

	param = &Param{ForceDarkMode: true}
	modeString = param.getMode().String()
	expected = "dark"

	assert.Equal(t, modeString, expected)
}
