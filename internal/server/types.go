package server

import "net/http"

type TemplateParam struct {
	Title            string
	Body             string
	Host             string
	Reload           bool
	Mode             string
	ShowBrowseButton bool
	IsDirectoryIndex bool
	HasReadme        bool
	DirectoryTitle   string
	Files            []FileInfo
	FileTree         []FileTreeItem
	CurrentPath      string
	ParentPath       string
	BreadcrumbItems  []BreadcrumbItem
}

type Param struct {
	Filename                       string
	MarkdownMode                   bool
	Reload                         bool
	ForceLightMode                 bool
	ForceDarkMode                  bool
	AutoOpen                       bool
	UseStdin                       bool
	StdinContent                   string
	DirectoryListing               bool
	DirectoryListingShowExtensions string
	DirectoryListingTextExtensions string
	IsDirectoryMode                bool
	DirectoryPath                  string
	ReadmeFile                     string
}

type Server struct {
	Host string
	Port int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

type mdResponseJSON struct {
	HTML  string `json:"html"`
	Title string `json:"title"`
}

type FileInfo struct {
	Name  string
	Path  string
	Depth int
}

type FileTreeItem struct {
	Name     string
	Path     string
	IsDir    bool
	IsBinary bool
	Children []FileTreeItem
}

type BreadcrumbItem struct {
	Name      string
	Path      string
	IsCurrent bool
}
