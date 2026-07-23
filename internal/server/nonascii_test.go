package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

// newNonASCIIDirParam creates a directory-mode Param backed by a temporary
// directory that contains non-ASCII and special-character filenames.
func newNonASCIIDirParam(t *testing.T) *Param {
	t.Helper()

	dir := t.TempDir()

	files := map[string]string{
		"日本語.md":     "# 日本語テスト\n\nHello from Japanese file.",
		"中文.md":      "# 中文测试",
		"foo&bar.md": "# Foo and Bar",
		"file#1.md":  "# File number one",
		"normal.md":  "# Normal file",
	}
	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)
		assert.Nil(t, err)
	}

	root, err := os.OpenRoot(dir)
	assert.Nil(t, err)

	t.Cleanup(func() { assert.Nil(t, root.Close()) })

	return &Param{
		DirectoryListing:               true,
		DirectoryListingShowExtensions: ".md",
		DirectoryListingTextExtensions: ".md,.txt",
		IsDirectoryMode:                true,
		DirectoryPath:                  dir,
		DirectoryRoot:                  root,
		Reload:                         false,
	}
}

// TestMdHandlerNonASCIIPath verifies the /__/md endpoint returns correct
// content when the path query parameter contains non-ASCII characters.
func TestMdHandlerNonASCIIPath(t *testing.T) {
	param := newNonASCIIDirParam(t)

	tests := []struct {
		name       string
		path       string
		wantTitle  string
		wantInHTML string
	}{
		{"Japanese filename", "日本語.md", "日本語.md", "日本語テスト"},
		{"Chinese filename", "中文.md", "中文.md", "中文测试"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/__/md?path="+url.QueryEscape(tt.path),
				nil,
			)
			rec := httptest.NewRecorder()

			mdHandler("", param).ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, http.StatusOK)

			body, err := io.ReadAll(res.Body)
			assert.Nil(t, err)

			var payload mdResponseJSON
			err = json.Unmarshal(body, &payload)
			assert.Nil(t, err)

			assert.Equal(t, payload.Title, tt.wantTitle)
			assert.True(t, strings.Contains(payload.HTML, tt.wantInHTML))
		})
	}
}

// TestDirectoryModeNonASCIIPageRender verifies that navigating to a page
// with a non-ASCII filename (percent-encoded URL) returns 200 and renders.
func TestDirectoryModeNonASCIIPageRender(t *testing.T) {
	param := newNonASCIIDirParam(t)

	w, err := watcher.Init(param.DirectoryPath)
	assert.Nil(t, err)

	defer w.Close()

	h := handler("", param, http.FileServer(http.Dir(param.DirectoryPath)), w)

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantIn   string
	}{
		{
			"Japanese file",
			"/" + url.PathEscape("日本語.md"),
			http.StatusOK,
			"日本語テスト",
		},
		{
			"Chinese file",
			"/" + url.PathEscape("中文.md"),
			http.StatusOK,
			"中文测试",
		},
		{
			"Ampersand file",
			"/" + url.PathEscape("foo&bar.md"),
			http.StatusOK,
			"Foo and Bar",
		},
		{
			"Hash file",
			"/" + url.PathEscape("file#1.md"),
			http.StatusOK,
			"File number one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, tt.wantCode)

			body, err := io.ReadAll(res.Body)
			assert.Nil(t, err)

			assert.True(t, strings.Contains(string(body), tt.wantIn))
		})
	}
}

// TestDirectoryListingHTMLEscapesNames verifies that filenames containing
// HTML-special characters (&, <, >) are escaped in the rendered HTML so
// they cannot break the markup or inject elements.
func TestDirectoryListingHTMLEscapesNames(t *testing.T) {
	param := newNonASCIIDirParam(t)

	w, err := watcher.Init(param.DirectoryPath)
	assert.Nil(t, err)

	defer w.Close()

	h := handler("", param, http.FileServer(http.Dir(param.DirectoryPath)), w)

	req := httptest.NewRequest(http.MethodGet, "/?view=index", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)

	htmlStr := string(body)

	// The display name "foo&bar.md" must appear HTML-escaped.
	assert.True(t, strings.Contains(htmlStr, "foo&amp;bar.md"))
	// Raw unescaped "&bar" must NOT appear outside of a proper entity.
	assert.False(t, strings.Contains(htmlStr, "foo&bar.md"))
}

// TestDirectoryListingURLEncodesHrefs verifies that href attributes in the
// directory listing are percent-encoded so that non-ASCII and URL-special
// characters (#, &) produce valid, clickable links.
func TestDirectoryListingURLEncodesHrefs(t *testing.T) {
	param := newNonASCIIDirParam(t)

	w, err := watcher.Init(param.DirectoryPath)
	assert.Nil(t, err)

	defer w.Close()

	h := handler("", param, http.FileServer(http.Dir(param.DirectoryPath)), w)

	req := httptest.NewRequest(http.MethodGet, "/?view=index", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)

	htmlStr := string(body)

	// Non-ASCII path must be percent-encoded in href.
	assert.False(t, strings.Contains(htmlStr, `href="/日本語.md"`))
	assert.True(t, strings.Contains(htmlStr, `href="/`+url.PathEscape("日本語.md")+`"`))

	// "#" must be encoded so the browser does not treat it as a fragment.
	assert.False(t, strings.Contains(htmlStr, `href="/file#1.md"`))
	assert.True(t, strings.Contains(htmlStr, `href="/`+url.PathEscape("file#1.md")+`"`))

	// "&" in href is HTML-escaped to &amp; by html/template.
	// The browser decodes &amp; back to & when following the link.
	assert.False(t, strings.Contains(htmlStr, `href="/foo&bar.md"`))
	assert.True(t, strings.Contains(htmlStr, `href="/foo&amp;bar.md"`))
}

// TestTemplateBodyNotEscaped verifies that the rendered Markdown body
// (which is intentionally raw HTML) is NOT double-escaped by the template.
func TestTemplateBodyNotEscaped(t *testing.T) {
	param := newNonASCIIDirParam(t)

	w, err := watcher.Init(param.DirectoryPath)
	assert.Nil(t, err)

	defer w.Close()

	h := handler("", param, http.FileServer(http.Dir(param.DirectoryPath)), w)

	req := httptest.NewRequest(http.MethodGet, "/"+url.PathEscape("日本語.md"), nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)

	htmlStr := string(body)

	// The rendered <h1> from Markdown must appear as real HTML, not escaped.
	assert.True(t, strings.Contains(htmlStr, "<h1"))
	assert.True(t, strings.Contains(htmlStr, "日本語テスト"))
	// Must NOT contain the escaped version.
	assert.False(t, strings.Contains(htmlStr, "&lt;h1"))
}

// TestBreadcrumbNonASCII verifies breadcrumb links are properly encoded.
func TestBreadcrumbNonASCII(t *testing.T) {
	param := newNonASCIIDirParam(t)

	w, err := watcher.Init(param.DirectoryPath)
	assert.Nil(t, err)

	defer w.Close()

	h := handler("", param, http.FileServer(http.Dir(param.DirectoryPath)), w)

	req := httptest.NewRequest(http.MethodGet, "/"+url.PathEscape("日本語.md"), nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)

	htmlStr := string(body)

	// Breadcrumb display name should be present (HTML-escaped if needed).
	assert.True(t, strings.Contains(htmlStr, "日本語.md"))
}
