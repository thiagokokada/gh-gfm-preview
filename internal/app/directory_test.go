package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
)

func TestParseExtensions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"Empty string", "", []string{".md"}},
		{"Single extension with dot", ".txt", []string{".txt"}},
		{"Single extension without dot", "txt", []string{".txt"}},
		{"Multiple extensions", ".md,.txt,.rst", []string{".md", ".txt", ".rst"}},
		{"Multiple extensions without dots", "md,txt,rst", []string{".md", ".txt", ".rst"}},
		{"Wildcard", "*", []string{"*"}},
		{"Wildcard with spaces", " * ", []string{"*"}},
		{"Wildcard in middle", ".txt,*,.md", []string{"*"}},
		{"Wildcard at end", ".txt,.md,*", []string{"*"}},
		{"Mixed case", ".MD,.TxT", []string{".md", ".txt"}},
		{"With spaces", " .md , .txt ", []string{".md", ".txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseExtensions(tt.input)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestListMarkdownFiles(t *testing.T) {
	tmp := t.TempDir()

	files := []string{
		"a.md",
		"b.txt",
		"c.MD",
		"sub/d.md",
		"sub/e.TXT",
	}

	// Create the directory tree
	for _, f := range files {
		full := filepath.Join(tmp, f)
		err := os.MkdirAll(filepath.Dir(full), 0o700)
		assert.Nil(t, err)

		err = os.WriteFile(full, []byte("x"), 0o600)
		assert.Nil(t, err)
	}

	t.Run("only .md", func(t *testing.T) {
		got, err := ListMarkdownFiles(tmp, []string{".md"})
		assert.Nil(t, err)

		want := []string{"a.md", "c.MD", filepath.Join("sub", "d.md")}
		assert.DeepEqual(t, got, want)
	})

	t.Run("multiple extensions", func(t *testing.T) {
		got, err := ListMarkdownFiles(tmp, []string{".md", ".txt"})
		assert.Nil(t, err)

		want := []string{
			"a.md", "b.txt", "c.MD",
			filepath.Join("sub", "d.md"),
			filepath.Join("sub", "e.TXT"),
		}
		assert.DeepEqual(t, got, want)
	})

	t.Run("wildcard returns everything", func(t *testing.T) {
		got, err := ListMarkdownFiles(tmp, []string{"*"})
		assert.Nil(t, err)

		want := []string{
			"a.md", "b.txt", "c.MD",
			filepath.Join("sub", "d.md"),
			filepath.Join("sub", "e.TXT"),
		}
		assert.DeepEqual(t, got, want)
	})
}

func TestListDirectoryContents(t *testing.T) {
	tmp := t.TempDir()

	// Create files + dirs
	err := errors.Join(
		os.WriteFile(filepath.Join(tmp, "a.md"), []byte("x"), 0o600),
		os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("x"), 0o600),
		os.WriteFile(filepath.Join(tmp, "c.html"), []byte("x"), 0o600),
		os.Mkdir(filepath.Join(tmp, "sub1"), 0o700),
		os.Mkdir(filepath.Join(tmp, "sub2"), 0o700),
	)
	assert.Nil(t, err)

	t.Run("only .md", func(t *testing.T) {
		files, dirs, err := ListDirectoryContents(tmp, []string{".md"})
		assert.Nil(t, err)

		wantFiles := []string{"a.md"}
		wantDirs := []string{"sub1", "sub2"}

		assert.DeepEqual(t, files, wantFiles)
		assert.DeepEqual(t, dirs, wantDirs)
	})

	t.Run("multiple extensions", func(t *testing.T) {
		files, _, err := ListDirectoryContents(tmp, []string{".md", ".txt"})
		assert.Nil(t, err)

		want := []string{"a.md", "b.txt"}
		assert.DeepEqual(t, files, want)
	})

	t.Run("wildcard", func(t *testing.T) {
		files, dirs, err := ListDirectoryContents(tmp, []string{"*"})
		assert.Nil(t, err)

		wantFiles := []string{"a.md", "b.txt", "c.html"}
		wantDirs := []string{"sub1", "sub2"}

		assert.DeepEqual(t, files, wantFiles)
		assert.DeepEqual(t, dirs, wantDirs)
	})
}

func TestIsTextFile(t *testing.T) {
	exts := []string{".md", ".txt"}

	tests := []struct {
		name string
		file string
		want bool
	}{
		{"match lowercase", "a.md", true},
		{"match different case", "b.MD", true},
		{"match txt", "x.txt", true},
		{"non-match", "image.jpg", false},
		{"no ext", "README", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTextFile(tt.file, exts)
			assert.Equal(t, got, tt.want)
		})
	}
}
