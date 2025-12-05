package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

func TestHandler(t *testing.T) {
	filename := "../../testdata/markdown-demo.md"
	dir := filepath.Dir(filename)
	param := &Param{
		Reload: false,
	}

	watcher, _ := watcher.Init(dir)
	defer watcher.Close()
	ts := httptest.NewServer(handler(filename, param, http.FileServer(http.Dir(dir)), watcher))
	defer ts.Close()

	r1, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}
	defer r1.Body.Close()

	if r1.StatusCode != http.StatusOK {
		t.Errorf("server status error, got: %v", r1.StatusCode)
	}

	if r1.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("content type error, got: %s\n", r1.Header.Get("Content-Type"))
	}

	r2, err := http.Get(ts.URL + "/images/dinotocat.png")
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}
	defer r2.Body.Close()

	if r2.StatusCode != http.StatusOK {
		t.Errorf("server status error, got: %v", r1.StatusCode)
	}

	if r2.Header.Get("Content-Type") != "image/png" {
		t.Errorf("content type error, got: %s\n", r2.Header.Get("Content-Type"))
	}
}

func TestMdHandler(t *testing.T) {
	filename := "../../testdata/markdown-demo.md"

	ts := httptest.NewServer(mdHandler(filename, &Param{}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("server status error, got: %v", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("content type error, got: %s\n", res.Header.Get("Content-Type"))
	}
}

func TestWrapHandler(t *testing.T) {
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello")
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := newLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, r)
		statusCode := lrw.statusCode

		// XXX
		if statusCode != http.StatusOK {
			t.Errorf("logging response status code error, got: %v", statusCode)
		}
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("server status error, got: %v", res.StatusCode)
	}
}
