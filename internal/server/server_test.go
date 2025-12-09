package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
