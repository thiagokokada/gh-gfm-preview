package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
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

	ts := httptest.NewServer(handler("", param, http.FileServer(http.Dir(testDir))))
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("status code error for %s: got %v, want %v", tt.path, res.StatusCode, tt.wantStatus)
			}
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

	ts := httptest.NewServer(handler("", param, http.FileServer(http.Dir(testDir))))
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("status code error for %s: got %v, want %v", tt.path, res.StatusCode, tt.wantStatus)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			bodyStr := string(body)

			if tt.wantInBody != "" && !strings.Contains(bodyStr, tt.wantInBody) {
				t.Errorf("response body should contain %q for %s", tt.wantInBody, tt.path)
				t.Logf("Body length: %d", len(bodyStr))
				t.Logf("Body preview (first 2000 chars): %s", bodyStr[:min(2000, len(bodyStr))])
			}

			if tt.wantNotInBody != "" && strings.Contains(bodyStr, tt.wantNotInBody) {
				t.Errorf("response body should not contain %q for %s", tt.wantNotInBody, tt.path)
				t.Logf("Body length: %d", len(bodyStr))
				t.Logf("Body preview (first 2000 chars): %s", bodyStr[:min(2000, len(bodyStr))])
			}
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

	ts := httptest.NewServer(handler("", param, http.FileServer(http.Dir(testDir))))
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("status code error for %s: got %v, want %v", tt.path, res.StatusCode, tt.wantStatus)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			bodyStr := string(body)

			for _, want := range tt.wantInBody {
				if !strings.Contains(bodyStr, want) {
					t.Errorf("response body should contain %q for %s", want, tt.path)
					t.Logf("Body preview (first 500 chars): %s", bodyStr[:min(500, len(bodyStr))])
				}
			}

			for _, notWant := range tt.wantNotInBody {
				if strings.Contains(bodyStr, notWant) {
					t.Errorf("response body should not contain %q for %s", notWant, tt.path)
				}
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
			if len(got) != len(tt.want) {
				t.Errorf("generateFileTree() length = %v, want %v", len(got), len(tt.want))

				return
			}

			for i := range got {
				if got[i].Name != tt.want[i].Name || got[i].Path != tt.want[i].Path || got[i].IsDir != tt.want[i].IsDir {
					t.Errorf("generateFileTree()[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
