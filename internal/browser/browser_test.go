package browser

import (
	"errors"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
)

type mockFileReader struct {
	content string
	err     error
}

func (m mockFileReader) readFile(_ string) (string, error) {
	return m.content, m.err
}

func TestIsContainWSL(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{
			name:     "WSL Data",
			data:     "Linux version 4.19.128-microsoft-standard (WSL2)",
			expected: true,
		},
		{
			name:     "Non-WSL Data",
			data:     "Linux version 4.15.0-72-generic",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isContainWSL(tt.data)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

var errMock = errors.New("mock error")

func TestIsWSL(t *testing.T) {
	tests := []struct {
		name     string
		reader   fileReader
		expected bool
	}{
		{
			name:     "WSL Data",
			reader:   mockFileReader{content: "Linux version 4.19.128-microsoft-standard (WSL2)"},
			expected: true,
		},
		{
			name:     "Non-WSL Data",
			reader:   mockFileReader{content: "Linux version 4.15.0-72-generic"},
			expected: false,
		},
		{
			name:     "Mock error",
			reader:   mockFileReader{err: errMock},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, isWSLWithReader(tt.reader), tt.expected)
		})
	}
}
