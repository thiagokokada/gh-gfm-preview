package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

func TestDirectoryModeCanonicalTrailingSlash(t *testing.T) {
	param := newNonASCIIDirParam(t)
	assert.Nil(t, os.Mkdir(filepath.Join(param.DirectoryPath, "sub"), 0o750))

	tests := []struct {
		name     string
		url      string
		status   int
		location string
	}{
		{"dir without slash", "/sub", http.StatusMovedPermanently, "/sub/"},
		{"dir without slash keeps query", "/sub?view=index", http.StatusMovedPermanently, "/sub/?view=index"},
		{"dir with slash", "/sub/", http.StatusOK, ""},
		{"file with slash", "/normal.md/", http.StatusMovedPermanently, "/normal.md"},
		{"file without slash", "/normal.md", http.StatusOK, ""},
		{"root", "/", http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)

			handleDirectoryMode(rec, req, param, watcher.NewDisabled())

			assert.Equal(t, tt.status, rec.Code)
			assert.Equal(t, tt.location, rec.Header().Get("Location"))
		})
	}
}
