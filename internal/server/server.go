package server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/andybalholm/crlf"
	"github.com/thiagokokada/gh-gfm-preview/internal/app"
	"github.com/thiagokokada/gh-gfm-preview/internal/browser"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
	"golang.org/x/text/transform"
)

//go:generate go run _tools/generate-assets.go

//go:embed template.html
var htmlTemplate string

//go:embed static/*
var staticDir embed.FS

// templateFuncs provides custom functions available in the HTML template.
var templateFuncs = template.FuncMap{
	// urlPathEscape percent-encodes p as a URL path. EscapedPath preserves
	// slash separators while encoding non-ASCII characters and URL-special
	// characters such as "#" and "&".
	//
	// The return type is template.URL so that html/template marks the value
	// as a safe URL and does not double-escape the percent signs.
	"urlPathEscape": func(p string) template.URL {
		return template.URL((&url.URL{Path: p}).EscapedPath()) //nolint:gosec // G203: URL is constructed from an escaped path
	},
}

var tmpl = template.Must(template.New("HTML Template").Funcs(templateFuncs).Parse(htmlTemplate))

const defaultPort = 3333

var (
	rootNormalizer     = new(crlf.Normalize)
	errNoDirectoryRoot = errors.New("directory root is not initialized")
)

func (server *Server) resolvePort() int {
	if server.Port > 0 {
		return server.Port
	}

	return defaultPort
}

func resolveFileAndDir(param *Param) (string, string, error) {
	if param.UseStdin {
		return "", ".", nil
	}

	return resolveFileMode(param)
}

func resolveFileMode(param *Param) (string, string, error) {
	inputPath := param.Filename
	if inputPath == "" {
		inputPath = "."
	}

	info, statErr := os.Stat(inputPath)
	isDir := statErr == nil && info.IsDir()

	if isDir && param.DirectoryListing {
		return setupDirectoryMode(param, inputPath)
	}

	return setupFileMode(param, inputPath)
}

func setupDirectoryMode(param *Param, inputPath string) (string, string, error) {
	param.IsDirectoryMode = true
	param.DirectoryPath = inputPath

	readme, readmeErr := app.FindReadmeFS(os.DirFS(inputPath), ".")
	if readmeErr == nil {
		param.ReadmeFile = filepath.Join(inputPath, readme)

		return param.ReadmeFile, inputPath, nil
	}

	return "", inputPath, nil
}

func setupFileMode(param *Param, inputPath string) (string, string, error) {
	filename, err := app.TargetFile(param.Filename)
	if err != nil {
		if param.DirectoryListing && errors.Is(err, app.ErrFileNotFound) {
			param.IsDirectoryMode = true
			param.DirectoryPath = inputPath

			return "", inputPath, nil
		}

		return "", "", fmt.Errorf("target file error: %w", err)
	}

	return filename, filepath.Dir(filename), nil
}

func (server *Server) Serve(param *Param) error {
	host := server.Host
	port := server.resolvePort()

	filename, dir, err := resolveFileAndDir(param)
	if err != nil {
		return err
	}

	if param.IsDirectoryMode {
		root, rootErr := os.OpenRoot(param.DirectoryPath)
		if rootErr != nil {
			return fmt.Errorf("directory root open error: %w", rootErr)
		}
		defer root.Close()

		param.DirectoryRoot = root
	}

	// Get the static subdirectory from embed.FS
	staticFS, err := fs.Sub(staticDir, "static")
	if err != nil {
		return fmt.Errorf("failed to get static subdirectory: %w", err)
	}

	watchTarget := watcherTarget(dir)

	watcher := initWatcher(watchTarget, param)
	defer watcher.Close()

	serveMux := http.NewServeMux()
	serveMux.Handle("/", wrapHandler(handler(filename, param, http.FileServer(http.Dir(dir)), watcher)))
	serveMux.Handle("/static/", wrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))))
	serveMux.Handle("/__/md", wrapHandler(mdHandler(filename, param)))

	serveMux.Handle("/ws", wsHandler(watcher))

	listener, err := getTCPListener(host, port)
	if err != nil {
		return err
	}

	address := listener.Addr()
	url := fmt.Sprintf("http://%s/", address)

	slog.Info("Accepting connections", "url", url)

	if param.AutoOpen {
		slog.Info("Opening URL in your browser", "url", url)

		go func() {
			err := browser.OpenBrowser(url)
			if err != nil {
				slog.Error("Error while opening browser", "error", err)
			}
		}()
	}

	hs := &http.Server{
		Handler:      serveMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	err = hs.Serve(listener)
	if err != nil {
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}

func watcherTarget(dir string) string {
	return dir
}

func initWatcher(watchTarget string, param *Param) *watcher.Watcher {
	w, err := watcher.Init(watchTarget)
	if err == nil {
		return w
	}

	slog.Warn("File watcher unavailable, live reload disabled", "path", watchTarget, "error", err)

	param.Reload = false

	return watcher.NewDisabled()
}

func handler(filename string, param *Param, handler http.Handler, watcher *watcher.Watcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !param.IsDirectoryMode {
			// Original single-file mode
			if !strings.HasSuffix(r.URL.Path, ".md") && r.URL.Path != "/" {
				handler.ServeHTTP(w, r)

				return
			}

			markdownView := mdResponse(w, filename, param)

			templateParam := TemplateParam{
				Title:        getTitle(filename),
				Body:         template.HTML(markdownView.HTML),         //nolint:gosec // G203: rendered Markdown is intentionally raw HTML
				HeadingsHTML: template.HTML(markdownView.HeadingsHTML), //nolint:gosec // G203: rendered Markdown is intentionally raw HTML
				HasHeadings:  markdownView.HasHeadings,
				Host:         r.Host,
				Reload:       param.Reload,
				Mode:         param.getMode().String(),
			}

			renderTemplate(w, templateParam)

			return
		}

		// Directory browsing mode
		handleDirectoryMode(w, r, param, watcher)
	})
}

func renderTemplate(w http.ResponseWriter, templateParam TemplateParam) {
	err := tmpl.Execute(w, templateParam)
	if err != nil {
		slog.Error("Template execute error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getMarkdown(filename string, param *Param) (string, error) {
	if param.UseStdin && param.StdinContent != "" && filename == "" {
		return param.StdinContent, nil
	}

	markdown, err := app.Slurp(filename)
	if err != nil {
		return "", fmt.Errorf("get markdown error: %w", err)
	}

	return markdown, nil
}

func mdResponse(w http.ResponseWriter, filename string, param *Param) markdownView {
	markdown, err := getMarkdown(filename, param)
	if err != nil {
		return writeMarkdownReadError(w, err)
	}

	return writeMarkdownViewResponse(w, markdown, param)
}

func writeMarkdownViewResponse(w http.ResponseWriter, markdown string, param *Param) markdownView {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	markdownView, err := renderMarkdownView(markdown, param)
	if err != nil {
		return writeMarkdownRenderError(w, err, param)
	}

	return markdownView
}

func renderMarkdownView(markdown string, param *Param) (markdownView, error) {
	html, err := app.ToHTML(markdown, param.MarkdownMode)
	if err != nil {
		return markdownView{}, fmt.Errorf("markdown convert error: %w", err)
	}

	headingsHTML, hasHeadings := renderHeadingsHTML(html)

	return markdownView{
		HTML:         html,
		HeadingsHTML: headingsHTML,
		HasHeadings:  hasHeadings,
	}, nil
}

func writeMarkdownReadError(w http.ResponseWriter, err error) markdownView {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	slog.Error("Error while reading markdown", "error", err)
	writeMarkdownError(w, err)

	return markdownView{}
}

func writeMarkdownRenderError(w http.ResponseWriter, err error, param *Param) markdownView {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	slog.Error(
		"Error while converting markdown to HTML",
		"mode", param.MarkdownMode,
		"error", err,
	)
	http.Error(w, err.Error(), http.StatusInternalServerError)

	return markdownView{}
}

func mdHandler(filename string, param *Param) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathParam := r.URL.Query().Get("path")

		if pathParam != "" && param.IsDirectoryMode {
			markdownView, title, err := renderMarkdownFromRoot(pathParam, param)
			if err != nil {
				writeMarkdownJSONErrorResponse(w, err, path.Base(pathParam))

				return
			}

			writeMarkdownJSONResponse(w, markdownView, title)

			return
		}

		if pathParam != "" {
			root, err := os.OpenRoot(filepath.Dir(filename))
			if err != nil {
				_ = writeMarkdownReadError(w, fmt.Errorf("directory root open error: %w", err))

				return
			}
			defer root.Close()

			markdownView, title, err := mdResponseFromOpenedRoot(w, pathParam, root, param)
			if err != nil {
				return
			}

			writeMarkdownJSONResponse(w, markdownView, title)

			return
		}

		handleSingleFileMarkdownRequest(w, filename, param)
	})
}

func handleSingleFileMarkdownRequest(w http.ResponseWriter, filename string, param *Param) {
	// If the file is a directory, try to find a README file
	if info, err := os.Stat(filename); err == nil && info.IsDir() {
		readme, err := app.FindReadmeFS(os.DirFS(filename), ".")
		if err == nil {
			filename = filepath.Join(filename, readme)
		}
	}

	title := getTitle(filename)

	markdown, err := getMarkdown(filename, param)
	if err != nil {
		writeMarkdownJSONErrorResponse(w, err, title)

		return
	}

	markdownView, err := renderMarkdownView(markdown, param)
	if err != nil {
		writeMarkdownJSONErrorResponse(w, err, title)

		return
	}

	writeMarkdownJSONResponse(w, markdownView, title)
}

func mdResponseFromRoot(w http.ResponseWriter, pathParam string, param *Param) (markdownView, string, error) {
	return mdResponseFromOpenedRoot(w, pathParam, param.DirectoryRoot, param)
}

func mdResponseFromOpenedRoot(w http.ResponseWriter, pathParam string, root *os.Root, param *Param) (markdownView, string, error) {
	markdownView, title, err := renderMarkdownFromOpenedRoot(pathParam, root, param)
	if err != nil {
		return writeMarkdownReadError(w, err), "", err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	return markdownView, title, nil
}

func renderMarkdownFromRoot(pathParam string, param *Param) (markdownView, string, error) {
	return renderMarkdownFromOpenedRoot(pathParam, param.DirectoryRoot, param)
}

func renderMarkdownFromOpenedRoot(pathParam string, root *os.Root, param *Param) (markdownView, string, error) {
	if root == nil {
		return markdownView{}, "", errNoDirectoryRoot
	}

	normalizedPath, ok := normalizeRootPath(pathParam)
	if !ok {
		err := fmt.Errorf("%w: %s", app.ErrFileNotFound, pathParam)

		return markdownView{}, "", err
	}

	file, title, err := resolveRootMarkdownTarget(root, normalizedPath)
	if err != nil {
		return markdownView{}, "", err
	}

	markdown, err := readRootMarkdown(root, file)
	if err != nil {
		return markdownView{}, "", err
	}

	view, err := renderMarkdownView(markdown, param)
	if err != nil {
		return markdownView{}, "", err
	}

	return view, title, nil
}

func writeMarkdownJSONResponse(w http.ResponseWriter, markdownView markdownView, title string) {
	w.Header().Set("Content-Type", "application/json")

	body, err := json.Marshal(mdResponseJSON{
		HTML:         markdownView.HTML,
		Title:        title,
		HeadingsHTML: markdownView.HeadingsHTML,
		HasHeadings:  markdownView.HasHeadings,
	})
	if err != nil {
		slog.Error("Error while JSON marshal", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	fmt.Fprintf(w, "%s", body)
}

func writeMarkdownJSONErrorResponse(w http.ResponseWriter, err error, title string) {
	slog.Error("Error while reading markdown", "error", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	writeMarkdownJSONResponse(w, markdownView{
		HTML: fmt.Sprintf("<pre>%s</pre>", html.EscapeString(err.Error())),
	}, title)
}

func resolveRootMarkdownTarget(root *os.Root, pathParam string) (string, string, error) {
	info, err := root.Stat(pathParam)
	if err == nil && info.IsDir() {
		readme, readmeErr := app.FindReadmeFS(root.FS(), rootRelativePath(pathParam))
		if readmeErr == nil {
			return readme, path.Base(readme), nil
		}
	}

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", fmt.Errorf("%w: %s", app.ErrFileNotFound, pathParam)
		}

		return "", "", fmt.Errorf("root stat error: %w", err)
	}

	return pathParam, path.Base(pathParam), nil
}

func normalizeRootPath(pathParam string) (string, bool) {
	if strings.HasPrefix(pathParam, "/") {
		return "", false
	}

	cleaned := path.Clean(pathParam)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false
	}

	if cleaned == "." {
		return ".", true
	}

	return cleaned, true
}

func readRootMarkdown(root *os.Root, path string) (string, error) {
	b, err := root.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%w: %s", app.ErrFileNotFound, path)
		}

		return "", fmt.Errorf("root read error: %w", err)
	}

	t, _, err := transform.Bytes(rootNormalizer, b)
	if err != nil {
		return "", fmt.Errorf("CRLF normalization error: %w", err)
	}

	return string(t), nil
}

func writeMarkdownError(w http.ResponseWriter, err error) {
	if errors.Is(err, app.ErrFileNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func wrapHandler(wrappedHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Disable cache, otherwise e.g., images will be cached and will not update
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		lrw := newLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, r)

		statusCode := lrw.statusCode
		slog.Debug("HTTP request", "method", r.Method, "code", statusCode, "url", r.URL)
	})
}

func getTitle(filename string) string {
	return filepath.Base(filename)
}

func getTCPListener(host string, port int) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		slog.Debug("Skipping port", "port", port, "error", err)
		listener, err = net.Listen("tcp", host+":0")
	}

	if err != nil {
		return nil, fmt.Errorf("TCP listener error: %w", err)
	}

	return listener, nil
}
