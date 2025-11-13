package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/thiagokokada/gh-gfm-preview/internal/app"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

// handleDirectoryMode handles HTTP requests in directory browsing mode
func handleDirectoryMode(w http.ResponseWriter, r *http.Request, param *Param) {
	// Directory mode - extract current path from URL
	urlPath := strings.TrimPrefix(r.URL.Path, "/")
	urlPath = strings.TrimSuffix(urlPath, "/")

	extensions := app.ParseExtensions(param.DirectoryListingShowExtensions)
	textExtensions := app.ParseExtensions(param.DirectoryListingTextExtensions)

	// Determine the actual filesystem path
	var currentDir string
	var currentURLPath string
	if urlPath == "" {
		currentDir = param.DirectoryPath
		currentURLPath = ""
	} else {
		currentDir = filepath.Join(param.DirectoryPath, urlPath)
		currentURLPath = urlPath
	}

	// Add this directory to the watcher if accessing it
	err := AddDirectoryToWatch(currentDir)
	if err != nil {
		utils.LogDebugf("Debug [failed to add directory to watcher]: %s: %v", currentDir, err)
	}

	// Security check: ensure currentDir is within param.DirectoryPath
	absBase, err := filepath.Abs(param.DirectoryPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	absCurrent, err := filepath.Abs(currentDir)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Ensure absBase ends with separator for proper prefix checking
	if !strings.HasSuffix(absBase, string(filepath.Separator)) {
		absBase += string(filepath.Separator)
	}

	// Check if current path is within base directory
	if absCurrent != strings.TrimSuffix(absBase, string(filepath.Separator)) && !strings.HasPrefix(absCurrent+string(filepath.Separator), absBase) {
		utils.LogDebugf("Path traversal attempt: base=%s, current=%s", absBase, absCurrent)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if it's a file request
	info, err := os.Stat(currentDir)
	isFile := err == nil && !info.IsDir()

	if isFile {
		// File access - also watch the parent directory
		fileDir := filepath.Dir(currentDir)
		err = AddDirectoryToWatch(fileDir)
		if err != nil {
			utils.LogDebugf("Debug [failed to add file's parent directory to watcher]: %s: %v", fileDir, err)
		}

		// File access - check if extension is allowed
		if !app.HasAllowedExtension(currentDir, extensions) {
			http.Error(w, "Forbidden: File type not allowed", http.StatusForbidden)
			return
		}

		// Check if file is a text file - if not, let browser handle it natively
		if !app.IsTextFile(currentDir, textExtensions) {
			// Serve file directly using http.ServeFile (binary file)
			http.ServeFile(w, r, currentDir)
			return
		}

		// Text file - render with markdown template
		templateParam := TemplateParam{
			Title:            getTitle(currentDir),
			Body:             mdResponse(w, currentDir, param),
			Host:             r.Host,
			Reload:           param.Reload,
			Mode:             param.getMode().String(),
			ShowBrowseButton: true,
			IsDirectoryIndex: false,
			HasReadme:        false,
			CurrentPath:      currentURLPath,
			ParentPath:       getParentPath(currentURLPath),
			BreadcrumbItems:  generateBreadcrumbItems(getParentPath(currentURLPath), filepath.Base(currentDir), false),
		}

		// Generate file tree for browse files popover (show files in the same directory)
		files, dirs, err := app.ListDirectoryContents(fileDir, extensions)
		if err == nil {
			// Use the directory path (parent of the file) for file tree
			dirURLPath := getParentPath(currentURLPath)
			templateParam.FileTree = generateFileTree(files, dirs, dirURLPath)
		}

		err = tmpl.Execute(w, templateParam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Directory access
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Check for README in current directory
	readme, readmeErr := app.FindReadme(currentDir)
	viewMode := r.URL.Query().Get("view")

	// Show directory listing
	if viewMode == "index" || readmeErr != nil {
		files, dirs, err := app.ListDirectoryContents(currentDir, extensions)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error listing directory: %v", err), http.StatusInternalServerError)
			return
		}

		// Determine directory title - use "Home" for root directory
		dirTitle := filepath.Base(currentURLPath)
		if currentURLPath == "" || dirTitle == "." {
			dirTitle = "Home"
		}

		templateParam := TemplateParam{
			Title:            "Browse Files",
			Body:             "",
			Host:             r.Host,
			Reload:           param.Reload,
			Mode:             param.getMode().String(),
			ShowBrowseButton: false,
			IsDirectoryIndex: true,
			HasReadme:        readmeErr == nil,
			DirectoryTitle:   dirTitle,
			FileTree:         generateFileTree(files, dirs, currentURLPath),
			CurrentPath:      currentURLPath,
			ParentPath:       getParentPath(currentURLPath),
			BreadcrumbItems:  generateBreadcrumbItems(getParentPath(currentURLPath), dirTitle, true),
		}

		err = tmpl.Execute(w, templateParam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Show README
	templateParam := TemplateParam{
		Title:            getTitle(readme),
		Body:             mdResponse(w, readme, param),
		Host:             r.Host,
		Reload:           param.Reload,
		Mode:             param.getMode().String(),
		ShowBrowseButton: true,
		IsDirectoryIndex: false,
		HasReadme:        true,
		CurrentPath:      currentURLPath,
		ParentPath:       getParentPath(currentURLPath),
		BreadcrumbItems:  generateBreadcrumbItems(currentURLPath, filepath.Base(readme), false),
	}

	// Generate file tree for popover
	files, dirs, err := app.ListDirectoryContents(currentDir, extensions)
	if err == nil {
		templateParam.FileTree = generateFileTree(files, dirs, currentURLPath)
	}

	err = tmpl.Execute(w, templateParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// generateFileTree creates FileTreeItem slice from files and directories
func generateFileTree(files []string, dirs []string, currentPath string) []FileTreeItem {
	items := make([]FileTreeItem, 0, len(dirs)+len(files)+1)

	// Add parent directory link (..) if not at root
	if currentPath != "" && currentPath != "." {
		parentPath := getParentPath(currentPath)
		items = append(items, FileTreeItem{
			Name:     "..",
			Path:     parentPath,
			IsDir:    true,
			IsBinary: false,
			Children: nil,
		})
	}

	// Add directories first
	for _, dir := range dirs {
		dirPath := filepath.Join(currentPath, dir)
		if currentPath == "" || currentPath == "." {
			dirPath = dir
		}
		items = append(items, FileTreeItem{
			Name:     dir,
			Path:     dirPath,
			IsDir:    true,
			IsBinary: false,
			Children: nil,
		})
	}

	// Add files
	for _, file := range files {
		filePath := filepath.Join(currentPath, file)
		if currentPath == "" || currentPath == "." {
			filePath = file
		}
		items = append(items, FileTreeItem{
			Name:     file,
			Path:     filePath,
			IsDir:    false,
			IsBinary: false, // Will be determined in template if needed
		})
	}

	return items
}

// getParentPath returns the parent path of the current path
func getParentPath(currentPath string) string {
	if currentPath == "" || currentPath == "." || currentPath == "/" {
		return ""
	}
	parent := filepath.Dir(currentPath)
	if parent == "." {
		return ""
	}
	return parent
}

// generateBreadcrumbItems creates breadcrumb items from current path and filename
// Returns all breadcrumb items including the current item
func generateBreadcrumbItems(currentPath, currentName string, isDirectory bool) []BreadcrumbItem {
	if currentPath == "" && currentName == "" {
		return []BreadcrumbItem{}
	}

	var items []BreadcrumbItem

	if currentPath != "" && currentPath != "." {
		parts := strings.Split(currentPath, string(filepath.Separator))
		currentBuildPath := ""

		for _, part := range parts {
			if part == "" {
				continue
			}
			if currentBuildPath == "" {
				currentBuildPath = part
			} else {
				currentBuildPath = filepath.Join(currentBuildPath, part)
			}
			items = append(items, BreadcrumbItem{
				Name:      part,
				Path:      currentBuildPath,
				IsCurrent: false,
			})
		}
	}

	// Add current file/directory if specified
	if currentName != "" {
		var path string
		if currentPath != "" && currentPath != "." {
			path = filepath.Join(currentPath, currentName)
		} else {
			path = currentName
		}

		items = append(items, BreadcrumbItem{
			Name:      currentName,
			Path:      path,
			IsCurrent: true,
		})
	}

	return items
}
