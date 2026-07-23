package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/thiagokokada/gh-gfm-preview/internal/app"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

// handleDirectoryMode handles HTTP requests in directory browsing mode.
func handleDirectoryMode(w http.ResponseWriter, r *http.Request, param *Param, watcher *watcher.Watcher) {
	urlPath := strings.TrimPrefix(r.URL.Path, "/")
	urlPath = strings.TrimSuffix(urlPath, "/")

	extensions := app.ParseExtensions(param.DirectoryListingShowExtensions)
	textExtensions := app.ParseExtensions(param.DirectoryListingTextExtensions)

	currentURLPath := resolveDirectoryPath(urlPath)

	info, err := statDirectoryTarget(param, currentURLPath)
	if err != nil {
		render404Error(w, r, param, currentURLPath)

		return
	}

	var currentHostPath string
	if info.IsDir() {
		currentHostPath = directoryHostPath(param.DirectoryPath, currentURLPath)
	} else {
		currentHostPath = directoryHostPath(param.DirectoryPath, getParentPath(currentURLPath))
	}

	err = watcher.AddDirectory(currentHostPath)
	if err != nil {
		slog.Debug("Add directory to watcher error", "error", err)
	}

	isFile := !info.IsDir()

	if isFile {
		handleFileRequest(w, r, param, watcher, currentURLPath, extensions, textExtensions, info)

		return
	}

	handleDirectoryRequest(w, r, param, currentURLPath, extensions)
}

func statDirectoryTarget(param *Param, currentURLPath string) (os.FileInfo, error) {
	if param.DirectoryRoot == nil {
		return nil, errNoDirectoryRoot
	}

	info, err := param.DirectoryRoot.Stat(rootRelativePath(currentURLPath))
	if err != nil {
		return nil, fmt.Errorf("directory target root stat error: %w", err)
	}

	return info, nil
}

func resolveDirectoryPath(urlPath string) string {
	return urlPath
}

func handleFileRequest(
	w http.ResponseWriter,
	r *http.Request,
	param *Param,
	watcher *watcher.Watcher,
	currentURLPath string,
	extensions []string,
	textExtensions []string,
	info os.FileInfo,
) {
	fileDirURLPath := getParentPath(currentURLPath)

	err := watcher.AddDirectory(directoryHostPath(param.DirectoryPath, fileDirURLPath))
	if err != nil {
		slog.Debug("Add directory to watcher error", "error", err)
	}

	if !app.IsTextFile(currentURLPath, textExtensions) {
		serveRootFile(w, r, param, currentURLPath, info)

		return
	}

	renderFileTemplate(w, r, param, currentURLPath, fileDirURLPath, extensions)
}

func renderFileTemplate(w http.ResponseWriter, r *http.Request, param *Param, currentURLPath, fileDirURLPath string, extensions []string) {
	markdownView, title, err := mdResponseFromRoot(w, currentURLPath, param)
	if err != nil {
		slog.Error("Error while reading markdown", "error", err)

		return
	}

	templateParam := TemplateParam{
		Title:            title,
		Body:             template.HTML(markdownView.HTML),
		HeadingsHTML:     template.HTML(markdownView.HeadingsHTML),
		HasHeadings:      markdownView.HasHeadings,
		Host:             r.Host,
		Reload:           param.Reload,
		Mode:             param.getMode().String(),
		ShowBrowseButton: true,
		IsDirectoryMode:  param.IsDirectoryMode,
		IsDirectoryIndex: false,
		HasReadme:        false,
		CurrentPath:      currentURLPath,
		ParentPath:       getParentPath(currentURLPath),
		BreadcrumbItems:  generateBreadcrumbItems(getParentPath(currentURLPath), path.Base(currentURLPath), false),
	}

	files, dirs, err := app.ListDirectoryContentsFS(param.DirectoryRoot.FS(), rootRelativePath(fileDirURLPath), extensions)
	if err == nil {
		dirURLPath := getParentPath(currentURLPath)
		templateParam.FileTree = generateFileTree(files, dirs, dirURLPath)
	}

	renderTemplate(w, templateParam)
}

func handleDirectoryRequest(w http.ResponseWriter, r *http.Request, param *Param, currentURLPath string, extensions []string) {
	readme, readmeErr := app.FindReadmeFS(param.DirectoryRoot.FS(), rootRelativePath(currentURLPath))
	viewMode := r.URL.Query().Get("view")

	if viewMode == "index" || readmeErr != nil {
		renderDirectoryListing(w, r, param, currentURLPath, extensions, readmeErr == nil)

		return
	}

	renderReadmeTemplate(w, r, param, currentURLPath, readme, extensions)
}

func renderDirectoryListing(w http.ResponseWriter, r *http.Request, param *Param, currentURLPath string, extensions []string, hasReadme bool) {
	files, dirs, err := app.ListDirectoryContentsFS(param.DirectoryRoot.FS(), rootRelativePath(currentURLPath), extensions)
	if err != nil {
		slog.Error("Error listing directory", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	dirTitle := path.Base(currentURLPath)
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
		IsDirectoryMode:  param.IsDirectoryMode,
		IsDirectoryIndex: true,
		HasReadme:        hasReadme,
		DirectoryTitle:   dirTitle,
		FileTree:         generateFileTree(files, dirs, currentURLPath),
		CurrentPath:      currentURLPath,
		ParentPath:       getParentPath(currentURLPath),
		BreadcrumbItems:  generateBreadcrumbItems(getParentPath(currentURLPath), dirTitle, true),
	}

	renderTemplate(w, templateParam)
}

func renderReadmeTemplate(w http.ResponseWriter, r *http.Request, param *Param, currentURLPath, readme string, extensions []string) {
	markdownView, title, err := mdResponseFromRoot(w, readme, param)
	if err != nil {
		slog.Error("Error while reading markdown", "error", err)

		return
	}

	templateParam := TemplateParam{
		Title:            title,
		Body:             template.HTML(markdownView.HTML),
		HeadingsHTML:     template.HTML(markdownView.HeadingsHTML),
		HasHeadings:      markdownView.HasHeadings,
		Host:             r.Host,
		Reload:           param.Reload,
		Mode:             param.getMode().String(),
		ShowBrowseButton: true,
		IsDirectoryMode:  param.IsDirectoryMode,
		IsDirectoryIndex: false,
		HasReadme:        true,
		CurrentPath:      currentURLPath,
		ParentPath:       getParentPath(currentURLPath),
		BreadcrumbItems:  generateBreadcrumbItems(currentURLPath, path.Base(readme), false),
	}

	files, dirs, err := app.ListDirectoryContentsFS(param.DirectoryRoot.FS(), rootRelativePath(currentURLPath), extensions)
	if err == nil {
		templateParam.FileTree = generateFileTree(files, dirs, currentURLPath)
	}

	renderTemplate(w, templateParam)
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
		dirPath := path.Join(currentPath, dir)
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
		filePath := path.Join(currentPath, file)
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

	parent := path.Dir(currentPath)
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

	parts := strings.Split(currentPath, "/")
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

	return path.Join(base, part)
}

func appendCurrentItem(items []BreadcrumbItem, currentPath, currentName string) []BreadcrumbItem {
	if currentName == "" {
		return items
	}

	itemPath := currentName
	if currentPath != "" && currentPath != "." {
		itemPath = path.Join(currentPath, currentName)
	}

	return append(items, BreadcrumbItem{
		Name:      currentName,
		Path:      itemPath,
		IsCurrent: true,
	})
}

func render404Error(w http.ResponseWriter, r *http.Request, param *Param, currentURLPath string) {
	w.WriteHeader(http.StatusNotFound)

	errorMessage := template.HTML("<h1>404 - Not Found</h1><p>The requested path does not exist.</p>")

	extensions := app.ParseExtensions(param.DirectoryListingShowExtensions)
	parentPath := getParentPath(currentURLPath)

	templateParam := TemplateParam{
		Title:            "404 - Not Found",
		Body:             errorMessage,
		Host:             r.Host,
		Reload:           param.Reload,
		Mode:             param.getMode().String(),
		ShowBrowseButton: true,
		BreadcrumbItems:  generateBreadcrumbItems(parentPath, path.Base(currentURLPath), false),
	}

	files, dirs, err := app.ListDirectoryContentsFS(param.DirectoryRoot.FS(), rootRelativePath(parentPath), extensions)
	if err == nil {
		templateParam.FileTree = generateFileTree(files, dirs, parentPath)
	}

	renderTemplate(w, templateParam)
}

func rootRelativePath(path string) string {
	if path == "" {
		return "."
	}

	return path
}

func directoryHostPath(basePath, relPath string) string {
	if relPath == "" {
		return basePath
	}

	return filepath.Join(basePath, relPath)
}

func serveRootFile(w http.ResponseWriter, r *http.Request, param *Param, currentURLPath string, info os.FileInfo) {
	file, err := param.DirectoryRoot.Open(currentURLPath)
	if err != nil {
		slog.Error("Error opening file from root", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer file.Close()

	http.ServeContent(w, r, path.Base(currentURLPath), info.ModTime(), file)
}
