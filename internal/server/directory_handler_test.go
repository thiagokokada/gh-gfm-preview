package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestDirectoryBrowsingMode(t *testing.T) {
	testDir := "../../testdata"
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
