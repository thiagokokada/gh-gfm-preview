package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/thiagokokada/gh-gfm-preview/internal/app"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

// handleDirectoryMode handles HTTP requests in directory browsing mode.
func handleDirectoryMode(w http.ResponseWriter, r *http.Request, param *Param, watcher *watcher.Watcher) {
	urlPath := strings.TrimPrefix(r.URL.Path, "/")
	urlPath = strings.TrimSuffix(urlPath, "/")

	extensions := app.ParseExtensions(param.DirectoryListingShowExtensions)
	textExtensions := app.ParseExtensions(param.DirectoryListingTextExtensions)

	currentDir, currentURLPath := resolveDirectoryPath(param.DirectoryPath, urlPath)

	err := watcher.AddDirectory(currentDir)
	if err != nil {
		utils.LogDebugf("Debug [add directory to watcher error]: %v", err)
	}

	if !validateDirectoryAccess(w, param.DirectoryPath, currentDir) {
		return
	}

	info, err := os.Stat(currentDir)
	isFile := err == nil && !info.IsDir()

	if isFile {
		handleFileRequest(w, r, param, watcher, currentDir, currentURLPath, extensions, textExtensions)

		return
	}

	if err != nil {
		render404Error(w, r, param, currentURLPath)

		return
	}

	handleDirectoryRequest(w, r, param, currentDir, currentURLPath, extensions)
}

func resolveDirectoryPath(basePath, urlPath string) (string, string) {
	if urlPath == "" {
		return basePath, ""
	}

	return filepath.Join(basePath, urlPath), urlPath
}

func validateDirectoryAccess(w http.ResponseWriter, basePath, currentPath string) bool {
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return false
	}

	absCurrent, err := filepath.Abs(currentPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return false
	}

	if !strings.HasSuffix(absBase, string(filepath.Separator)) {
		absBase += string(filepath.Separator)
	}

	isValid := absCurrent == strings.TrimSuffix(absBase, string(filepath.Separator)) ||
		strings.HasPrefix(absCurrent+string(filepath.Separator), absBase)

	if !isValid {
		utils.LogDebugf("Path traversal attempt: base=%s, current=%s", absBase, absCurrent)
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	return isValid
}

func handleFileRequest(
	w http.ResponseWriter,
	r *http.Request,
	param *Param,
	watcher *watcher.Watcher,
	currentDir string,
	currentURLPath string,
	extensions []string,
	textExtensions []string,
) {
	fileDir := filepath.Dir(currentDir)

	err := watcher.AddDirectory(fileDir)
	if err != nil {
		utils.LogDebugf("Debug [add directory to watcher error]: %v", err)
	}

	if !app.HasAllowedExtension(currentDir, extensions) {
		http.Error(w, "Forbidden: File type not allowed", http.StatusForbidden)

		return
	}

	if !app.IsTextFile(currentDir, textExtensions) {
		http.ServeFile(w, r, currentDir)

		return
	}

	renderFileTemplate(w, r, param, currentDir, currentURLPath, fileDir, extensions)
}

func renderFileTemplate(w http.ResponseWriter, r *http.Request, param *Param, currentDir, currentURLPath, fileDir string, extensions []string) {
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

	files, dirs, err := app.ListDirectoryContents(fileDir, extensions)
	if err == nil {
		dirURLPath := getParentPath(currentURLPath)
		templateParam.FileTree = generateFileTree(files, dirs, dirURLPath)
	}

	err = tmpl.Execute(w, templateParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDirectoryRequest(w http.ResponseWriter, r *http.Request, param *Param, currentDir, currentURLPath string, extensions []string) {
	readme, readmeErr := app.FindReadme(currentDir)
	viewMode := r.URL.Query().Get("view")

	if viewMode == "index" || readmeErr != nil {
		renderDirectoryListing(w, r, param, currentDir, currentURLPath, extensions, readmeErr == nil)

		return
	}

	renderReadmeTemplate(w, r, param, currentDir, currentURLPath, readme, extensions)
}

func renderDirectoryListing(w http.ResponseWriter, r *http.Request, param *Param, currentDir, currentURLPath string, extensions []string, hasReadme bool) {
	files, dirs, err := app.ListDirectoryContents(currentDir, extensions)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing directory: %v", err), http.StatusInternalServerError)

		return
	}

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
		HasReadme:        hasReadme,
		DirectoryTitle:   dirTitle,
		FileTree:         generateFileTree(files, dirs, currentURLPath),
		CurrentPath:      currentURLPath,
		ParentPath:       getParentPath(currentURLPath),
		BreadcrumbItems:  generateBreadcrumbItems(getParentPath(currentURLPath), dirTitle, true),
	}

	err = tmpl.Execute(w, templateParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderReadmeTemplate(w http.ResponseWriter, r *http.Request, param *Param, currentDir, currentURLPath, readme string, extensions []string) {
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

	files, dirs, err := app.ListDirectoryContents(currentDir, extensions)
	if err == nil {
		templateParam.FileTree = generateFileTree(files, dirs, currentURLPath)
	}

	err = tmpl.Execute(w, templateParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// generateFileTree creates FileTreeItem slice from files and directories.
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

// getParentPath returns the parent path of the current path.
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
// Returns all breadcrumb items including the current item.
func generateBreadcrumbItems(currentPath, currentName string, _ bool) []BreadcrumbItem {
	if currentPath == "" && currentName == "" {
		return []BreadcrumbItem{}
	}

	items := buildPathBreadcrumbs(currentPath)
	items = appendCurrentItem(items, currentPath, currentName)

	return items
}

func buildPathBreadcrumbs(currentPath string) []BreadcrumbItem {
	if currentPath == "" || currentPath == "." {
		return []BreadcrumbItem{}
	}

	parts := strings.Split(currentPath, string(filepath.Separator))
	items := make([]BreadcrumbItem, 0, len(parts))
	currentBuildPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		currentBuildPath = buildPath(currentBuildPath, part)
		items = append(items, BreadcrumbItem{
			Name:      part,
			Path:      currentBuildPath,
			IsCurrent: false,
		})
	}

	return items
}

func buildPath(base, part string) string {
	if base == "" {
		return part
	}

	return filepath.Join(base, part)
}

func appendCurrentItem(items []BreadcrumbItem, currentPath, currentName string) []BreadcrumbItem {
	if currentName == "" {
		return items
	}

	path := currentName
	if currentPath != "" && currentPath != "." {
		path = filepath.Join(currentPath, currentName)
	}

	return append(items, BreadcrumbItem{
		Name:      currentName,
		Path:      path,
		IsCurrent: true,
	})
}

func render404Error(w http.ResponseWriter, r *http.Request, param *Param, currentURLPath string) {
	w.WriteHeader(http.StatusNotFound)

	errorMessage := "<h1>404 - Not Found</h1><p>The requested path does not exist.</p>"

	extensions := app.ParseExtensions(param.DirectoryListingShowExtensions)
	parentPath := getParentPath(currentURLPath)
	parentDir := filepath.Join(param.DirectoryPath, parentPath)

	templateParam := TemplateParam{
		Title:            "404 - Not Found",
		Body:             errorMessage,
		Host:             r.Host,
		Reload:           param.Reload,
		Mode:             param.getMode().String(),
		ShowBrowseButton: true,
		BreadcrumbItems:  generateBreadcrumbItems(parentPath, filepath.Base(currentURLPath), false),
	}

	files, dirs, err := app.ListDirectoryContents(parentDir, extensions)
	if err == nil {
		templateParam.FileTree = generateFileTree(files, dirs, parentPath)
	}

	err = tmpl.Execute(w, templateParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
