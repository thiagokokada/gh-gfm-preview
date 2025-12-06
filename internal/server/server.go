package server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/thiagokokada/gh-gfm-preview/internal/app"
	"github.com/thiagokokada/gh-gfm-preview/internal/browser"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

//go:generate go run _tools/generate-assets.go

//go:embed template.html
var htmlTemplate string

//go:embed static/*
var staticDir embed.FS
var tmpl = template.Must(template.New("HTML Template").Parse(htmlTemplate))

const defaultPort = 3333

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

	readme, readmeErr := app.FindReadme(inputPath)
	if readmeErr == nil {
		param.ReadmeFile = readme

		return readme, inputPath, nil
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

	// Get the static subdirectory from embed.FS
	staticFS, err := fs.Sub(staticDir, "static")
	if err != nil {
		return fmt.Errorf("failed to get static subdirectory: %w", err)
	}

	watcher, err := watcher.Init(dir)
	if err != nil {
		return fmt.Errorf("error while file watcher init: %w", err)
	}
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

func handler(filename string, param *Param, handler http.Handler, watcher *watcher.Watcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !param.IsDirectoryMode {
			// Original single-file mode
			if !strings.HasSuffix(r.URL.Path, ".md") && r.URL.Path != "/" {
				handler.ServeHTTP(w, r)

				return
			}

			templateParam := TemplateParam{
				Title:  getTitle(filename),
				Body:   mdResponse(w, filename, param),
				Host:   r.Host,
				Reload: param.Reload,
				Mode:   param.getMode().String(),
			}

			err := tmpl.Execute(w, templateParam)
			if err != nil {
				slog.Error("Template execute error", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			return
		}

		// Directory browsing mode
		handleDirectoryMode(w, r, param, watcher)
	})
}

func getMarkdown(filename string, param *Param) (string, error) {
	if param.UseStdin && param.StdinContent != "" && filename == "" {
		return param.StdinContent, nil
	}

	markdown, err := app.Slurp(filename)
	if err != nil {
		slog.Error("Error while reading markdown", "error", err)

		return "", fmt.Errorf("error while reading markdown: %w", err)
	}

	return markdown, nil
}

func mdResponse(w http.ResponseWriter, filename string, param *Param) string {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	markdown, err := getMarkdown(filename, param)
	if err != nil {
		slog.Error("Error while reading markdown", "error", err)

		if errors.Is(err, app.ErrFileNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return ""
	}

	html, err := app.ToHTML(markdown, param.MarkdownMode)
	if err != nil {
		slog.Error(
			"Error while converting markdown to HTML",
			"mode", param.MarkdownMode,
			"error", err,
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return ""
	}

	return html
}

func mdHandler(filename string, param *Param) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathParam := r.URL.Query().Get("path")

		var file string

		if pathParam != "" {
			// In directory mode, convert relative path to absolute path
			if param.IsDirectoryMode {
				file = filepath.Join(param.DirectoryPath, pathParam)
			} else {
				file = pathParam
			}
		} else {
			file = filename
		}

		// If the file is a directory, try to find a README file
		if info, err := os.Stat(file); err == nil && info.IsDir() {
			readme, err := app.FindReadme(file)
			if err == nil {
				file = readme
			}
		}

		html := mdResponse(w, file, param)
		title := getTitle(file)

		body, err := json.Marshal(mdResponseJSON{HTML: html, Title: title})
		if err != nil {
			slog.Error("Error while JSON marshal", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		fmt.Fprintf(w, "%s", body)
	})
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
