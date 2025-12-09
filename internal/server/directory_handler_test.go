package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

const testDataDir = "../../testdata"

func TestDirectoryBrowsingMode(t *testing.T) {
	testDir := testDataDir
	param := &Param{
		DirectoryListing:               true,
		DirectoryListingShowExtensions: ".md",
		DirectoryListingTextExtensions: ".md,.txt",
		IsDirectoryMode:                true,
		DirectoryPath:                  testDir,
		ReadmeFile:                     filepath.Join(testDir, "README"),
		Reload:                         false,
	}

	watcher, err := watcher.Init(testDir)
	assert.Nil(t, err)

	defer watcher.Close()

	ts := httptest.NewServer(handler("", param, http.FileServer(http.Dir(testDir)), watcher))
	defer ts.Close()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"Root README", "/", http.StatusOK},
		{"Markdown file", "/markdown-demo.md", http.StatusOK},
		{"Non-existent file", "/does-not-exist.md", http.StatusNotFound},
		{"Directory listing", "/?view=index", http.StatusOK},
		{"Subdirectory access", "/images/", http.StatusOK},
		{"Subdirectory README with trailing slash", "/subdir/", http.StatusOK},
		{"Subdirectory README.md explicit", "/subdir/README.md", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := http.Get(ts.URL + tt.path)
			assert.Nil(t, err)

			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, tt.wantStatus)
		})
	}
}

func TestSubdirectoryReadmeAccess(t *testing.T) {
	testDir := testDataDir
	param := &Param{
		DirectoryListing:               true,
		DirectoryListingShowExtensions: ".md",
		DirectoryListingTextExtensions: ".md,.txt",
		IsDirectoryMode:                true,
		DirectoryPath:                  testDir,
		ReadmeFile:                     filepath.Join(testDir, "README"),
		Reload:                         false,
	}

	watcher, err := watcher.Init(testDir)
	assert.Nil(t, err)

	defer watcher.Close()

	ts := httptest.NewServer(handler("", param, http.FileServer(http.Dir(testDir)), watcher))
	defer ts.Close()

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantInBody    string
		wantNotInBody string
	}{
		{
			name:          "Subdirectory with trailing slash should show README",
			path:          "/subdir/",
			wantStatus:    http.StatusOK,
			wantInBody:    "Subdirectory README",
			wantNotInBody: "directory-index markdown-body",
		},
		{
			name:          "Subdirectory README.md explicit",
			path:          "/subdir/README.md",
			wantStatus:    http.StatusOK,
			wantInBody:    "Subdirectory README",
			wantNotInBody: "directory-index markdown-body",
		},
		{
			name:          "Subdirectory without trailing slash should show README",
			path:          "/subdir",
			wantStatus:    http.StatusOK,
			wantInBody:    "Subdirectory README",
			wantNotInBody: "directory-index markdown-body",
		},
		{
			name:       "Subdirectory with view=index should show directory listing",
			path:       "/subdir/?view=index",
			wantStatus: http.StatusOK,
			wantInBody: "directory-index markdown-body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := http.Get(ts.URL + tt.path)
			assert.Nil(t, err)

			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, tt.wantStatus)

			body, err := io.ReadAll(res.Body)
			assert.Nil(t, err)

			bodyStr := string(body)
			assert.True(t, tt.wantInBody != "" && strings.Contains(bodyStr, tt.wantInBody))
			assert.False(t, tt.wantNotInBody != "" && strings.Contains(bodyStr, tt.wantNotInBody))
		})
	}
}

func Test404ErrorRendering(t *testing.T) {
	testDir := testDataDir
	param := &Param{
		DirectoryListing:               true,
		DirectoryListingShowExtensions: ".md",
		DirectoryListingTextExtensions: ".md,.txt",
		IsDirectoryMode:                true,
		DirectoryPath:                  testDir,
		ReadmeFile:                     filepath.Join(testDir, "README"),
		Reload:                         false,
	}

	watcher, _ := watcher.Init(testDir)
	defer watcher.Close()

	ts := httptest.NewServer(handler("", param, http.FileServer(http.Dir(testDir)), watcher))
	defer ts.Close()

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantInBody    []string
		wantNotInBody []string
	}{
		{
			name:       "Non-existent file should show 404 with template",
			path:       "/does-not-exist.md",
			wantStatus: http.StatusNotFound,
			wantInBody: []string{
				"404 - Not Found",
				"breadcrumb",
				"<html>",
			},
			wantNotInBody: []string{},
		},
		{
			name:       "Non-existent directory should show 404 with template",
			path:       "/nonexistent-dir/",
			wantStatus: http.StatusNotFound,
			wantInBody: []string{
				"404 - Not Found",
				"breadcrumb",
				"<html>",
			},
			wantNotInBody: []string{},
		},
		{
			name:       "Non-existent nested path should show 404 with template",
			path:       "/some/deep/path/file.md",
			wantStatus: http.StatusNotFound,
			wantInBody: []string{
				"404 - Not Found",
				"breadcrumb",
				"<html>",
			},
			wantNotInBody: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := http.Get(ts.URL + tt.path)
			assert.Nil(t, err)

			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, tt.wantStatus)

			body, err := io.ReadAll(res.Body)
			assert.Nil(t, err)

			bodyStr := string(body)

			for _, want := range tt.wantInBody {
				assert.True(t, strings.Contains(bodyStr, want))
			}

			for _, notWant := range tt.wantNotInBody {
				assert.False(t, strings.Contains(bodyStr, notWant))
			}
		})
	}
}

func TestGenerateFileTree(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		dirs        []string
		currentPath string
		want        []FileTreeItem
	}{
		{
			name:        "Root directory",
			files:       []string{"file1.md", "file2.md"},
			dirs:        []string{"dir1", "dir2"},
			currentPath: "",
			want: []FileTreeItem{
				{Name: "dir1", Path: "dir1", IsDir: true},
				{Name: "dir2", Path: "dir2", IsDir: true},
				{Name: "file1.md", Path: "file1.md", IsDir: false},
				{Name: "file2.md", Path: "file2.md", IsDir: false},
			},
		},
		{
			name:        "Subdirectory",
			files:       []string{"file.md"},
			dirs:        []string{"subdir"},
			currentPath: "parent",
			want: []FileTreeItem{
				{Name: "..", Path: "", IsDir: true},
				{Name: "subdir", Path: "parent/subdir", IsDir: true},
				{Name: "file.md", Path: "parent/file.md", IsDir: false},
			},
		},
		{
			name:        "Nested subdirectory",
			files:       []string{},
			dirs:        []string{"BB"},
			currentPath: "AA",
			want: []FileTreeItem{
				{Name: "..", Path: "", IsDir: true},
				{Name: "BB", Path: "AA/BB", IsDir: true},
			},
		},
		{
			name:        "Deeply nested subdirectory",
			files:       []string{"file.md"},
			dirs:        []string{"CC"},
			currentPath: "AA/BB",
			want: []FileTreeItem{
				{Name: "..", Path: "AA", IsDir: true},
				{Name: "CC", Path: "AA/BB/CC", IsDir: true},
				{Name: "file.md", Path: "AA/BB/file.md", IsDir: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFileTree(tt.files, tt.dirs, tt.currentPath)
			assert.Equal(t, len(got), len(tt.want))

			for i := range got {
				assert.Equal(t, got[i].Name, tt.want[i].Name)
				assert.Equal(t, got[i].Path, tt.want[i].Path)
				assert.Equal(t, got[i].IsDir, tt.want[i].IsDir)
			}
		})
	}
}
