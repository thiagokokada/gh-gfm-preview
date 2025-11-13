package server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/thiagokokada/gh-gfm-preview/internal/app"
	"github.com/thiagokokada/gh-gfm-preview/internal/browser"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
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

	serveMux := http.NewServeMux()
	serveMux.Handle("/", wrapHandler(handler(filename, param, http.FileServer(http.Dir(dir)))))
	serveMux.Handle("/static/", wrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))))
	serveMux.Handle("/__/md", wrapHandler(mdHandler(filename, param)))

	watcher, err := createWatcher(dir)
	if err != nil {
		return err
	}
	defer watcher.Close()

	serveMux.Handle("/ws", wsHandler(watcher))

	listener, err := getTCPListener(host, port)
	if err != nil {
		return err
	}

	address := listener.Addr()

	utils.LogInfof("Accepting connections at http://%s/\n", address)

	if param.AutoOpen {
		utils.LogInfof("Open http://%s/ on your browser\n", address)

		go func() {
			err := browser.OpenBrowser(fmt.Sprintf("http://%s/", address))
			if err != nil {
				utils.LogInfof("Error while opening browser: %s\n", err)
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

func handler(filename string, param *Param, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !param.IsDirectoryMode {
			// Original single-file mode
			if !strings.HasSuffix(r.URL.Path, ".md") && r.URL.Path != "/" {
				h.ServeHTTP(w, r)

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
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			return
		}

		// Directory browsing mode
		handleDirectoryMode(w, r, param)
	})
}

func mdResponse(w http.ResponseWriter, filename string, param *Param) string {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var markdown string

	var err error

	if param.UseStdin && param.StdinContent != "" && filename == "" {
		markdown = param.StdinContent
	} else {
		markdown, err = app.Slurp(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return ""
		}
	}

	if err != nil {
		if errors.Is(err, app.ErrFileNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return ""
	}

	html, err := app.ToHTML(markdown, param.MarkdownMode)
	if err != nil {
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

		html := mdResponse(w, file, param)
		title := getTitle(file)

		body, err := json.Marshal(mdResponseJSON{HTML: html, Title: title})
		if err != nil {
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
		lrw := newLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, r)

		statusCode := lrw.statusCode
		utils.LogInfof("%s [%d] %s", r.Method, statusCode, r.URL)
	})
}

func getTitle(filename string) string {
	return filepath.Base(filename)
}

func getTCPListener(host string, port int) (net.Listener, error) {
	var err error

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		utils.LogInfof("Skipping port %d: %v", port, err)
		listener, err = net.Listen("tcp", host+":0")
	}

	if err != nil {
		return nil, fmt.Errorf("TCP listener error: %w", err)
	}

	return listener, nil
}
