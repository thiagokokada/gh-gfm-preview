package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

func TestHandler(t *testing.T) {
	filename := "../../testdata/markdown-demo.md"
	dir := filepath.Dir(filename)
	param := &Param{
		Reload: false,
	}

	watcher, err := watcher.Init(dir)
	assert.Nil(t, err)

	defer watcher.Close()

	ts := httptest.NewServer(handler(filename, param, http.FileServer(http.Dir(dir)), watcher))
	defer ts.Close()

	r1, err := http.Get(ts.URL)
	assert.Nil(t, err)

	defer r1.Body.Close()

	assert.Equal(t, r1.StatusCode, http.StatusOK)
	assert.Equal(t, r1.Header.Get("Content-Type"), "text/html; charset=utf-8")

	r2, err := http.Get(ts.URL + "/images/dinotocat.png")
	assert.Nil(t, err)

	defer r2.Body.Close()

	assert.Equal(t, r2.StatusCode, http.StatusOK)
	assert.Equal(t, r2.Header.Get("Content-Type"), "image/png")
}

func TestMdHandler(t *testing.T) {
	filename := "../../testdata/markdown-demo.md"

	ts := httptest.NewServer(mdHandler(filename, &Param{}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.Nil(t, err)

	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Equal(t, res.Header.Get("Content-Type"), "text/html; charset=utf-8")

	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)

	var payload mdResponseJSON
	err = json.Unmarshal(body, &payload)
	assert.Nil(t, err)

	assert.True(t, payload.HasHeadings)
	assert.True(t, strings.Contains(payload.HeadingsHTML, `class="heading-item heading-level-1"`))
	assert.True(t, strings.Contains(payload.HeadingsHTML, `href="#headings"`))
}

func TestMdHandlerWithoutHeadings(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "no-headings-*.md")
	assert.Nil(t, err)

	defer tmpFile.Close()

	_, err = tmpFile.WriteString("just plain text")
	assert.Nil(t, err)

	ts := httptest.NewServer(mdHandler(tmpFile.Name(), &Param{}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.Nil(t, err)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)

	var payload mdResponseJSON
	err = json.Unmarshal(body, &payload)
	assert.Nil(t, err)

	assert.False(t, payload.HasHeadings)
	assert.Equal(t, payload.HeadingsHTML, "")
}

func TestWrapHandler(t *testing.T) {
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello")
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := newLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, r)

		// XXX
		assert.Equal(t, lrw.statusCode, http.StatusOK)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.Nil(t, err)

	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)
}
