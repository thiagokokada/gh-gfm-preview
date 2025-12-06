package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ParseExtensions parses a comma-separated string of file extensions.
// Special case: if "*" is found anywhere in the list, returns []string{"*"} to match all files.
func ParseExtensions(extensionsStr string) []string {
	if extensionsStr == "" {
		return []string{".md"}
	}

	parts := strings.Split(extensionsStr, ",")
	extensions := make([]string, 0, len(parts))
	hasWildcard := false

	for _, ext := range parts {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}
		// Check for wildcard
		if ext == "*" {
			hasWildcard = true

			continue
		}
		// Add leading dot if not present
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}

		extensions = append(extensions, strings.ToLower(ext))
	}

	// If wildcard is present, return wildcard only
	if hasWildcard {
		return []string{"*"}
	}

	if len(extensions) == 0 {
		return []string{".md"}
	}

	return extensions
}

// ListMarkdownFiles recursively lists files with specified extensions in a directory.
func ListMarkdownFiles(dir string, extensions []string) ([]string, error) {
	var files []string

	// Check if wildcard is enabled
	allowAll := len(extensions) == 1 && extensions[0] == "*"

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// If wildcard, include all files
		if allowAll {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			files = append(files, relPath)

			return nil
		}

		// Check if file has one of the specified extensions (case-insensitive)
		ext := strings.ToLower(filepath.Ext(path))
		for _, validExt := range extensions {
			if ext == validExt {
				// Get relative path from dir
				relPath, err := filepath.Rel(dir, path)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}

				files = append(files, relPath)

				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	// Sort files alphabetically
	sort.Strings(files)

	return files, nil
}

// ListDirectoryContents lists only the immediate contents (files and directories) of a directory.
func ListDirectoryContents(dir string, extensions []string) ([]string, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading directory: %w", err)
	}

	var files, dirs []string

	// Check if wildcard is enabled
	allowAll := len(extensions) == 1 && extensions[0] == "*"

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		} else {
			// If wildcard, include all files
			if allowAll {
				files = append(files, entry.Name())

				continue
			}
			// Check if file has one of the specified extensions (case-insensitive)
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			for _, validExt := range extensions {
				if ext == validExt {
					files = append(files, entry.Name())

					break
				}
			}
		}
	}

	// Sort alphabetically
	sort.Strings(files)
	sort.Strings(dirs)

	return files, dirs, nil
}

// IsTextFile checks if a file is a text file based on allowed extensions (whitelist)
// Returns true if the file extension is in the allowed list.
func IsTextFile(filePath string, textExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}
