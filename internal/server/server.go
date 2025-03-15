package server

import (
	"embed"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	dark "github.com/thiagokokada/dark-mode-go"
	"github.com/thiagokokada/gh-gfm-preview/internal/app"
	"github.com/thiagokokada/gh-gfm-preview/internal/browser"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

//go:embed template.html
var htmlTemplate string

//go:embed static/*
var staticDir embed.FS
var tmpl = template.Must(template.New("HTML Template").Parse(htmlTemplate))

const (
	defaultPort = 3333
	darkMode    = "dark"
	lightMode   = "light"
	defaultMode = darkMode
)

func (server *Server) Serve(param *Param) error {
	host := server.Host

	port := defaultPort
	if server.Port > 0 {
		port = server.Port
	}

	filename, err := app.TargetFile(param.Filename)
	if err != nil {
		return fmt.Errorf("target file error: %w", err)
	}

	dir := filepath.Dir(filename)

	serveMux := http.NewServeMux()
	serveMux.Handle("/", wrapHandler(handler(filename, param, http.FileServer(http.Dir(dir)))))
	serveMux.Handle("/static/", wrapHandler(handler(filename, param, http.FileServer(http.FS(staticDir)))))
	serveMux.Handle("/__/md", wrapHandler(mdHandler(filename, param)))

	watcher, err := createWatcher(dir)
	if err != nil {
		return err
	}
	defer watcher.Close()

	serveMux.Handle("/ws", wsHandler(watcher))

	port, err = getPort(host, port)
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%d", host, port)

	utils.LogInfo("Accepting connections at http://%s/\n", address)

	if param.AutoOpen {
		utils.LogInfo("Open http://%s/ on your browser\n", address)

		go func() {
			err := browser.OpenBrowser(fmt.Sprintf("http://%s/", address))
			if err != nil {
				utils.LogInfo("Error while opening browser: %s\n", err)
			}
		}()
	}

	httpServer := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           serveMux,
	}

	err = httpServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("listen and serve error: %w", err)
	}

	return nil
}

func handler(filename string, param *Param, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".md") && r.URL.Path != "/" {
			h.ServeHTTP(w, r)

			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		markdown, err := app.Slurp(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		html, err := app.ToHTML(markdown, param.MarkdownMode, isDarkMode(param))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		param := TemplateParam{
			Title:  getTitle(filename),
			Body:   html,
			Host:   r.Host,
			Reload: param.Reload,
			Mode:   getMode(param),
		}

		err = tmpl.Execute(w, param)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	})
}

func mdResponse(w http.ResponseWriter, filename string, param *Param) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	markdown, err := app.Slurp(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	html, err := app.ToHTML(markdown, param.MarkdownMode, isDarkMode(param))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	fmt.Fprintf(w, "%s", html)
}

func mdHandler(filename string, param *Param) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathParam := r.URL.Query().Get("path")
		if pathParam != "" {
			mdResponse(w, pathParam, param)
		} else {
			mdResponse(w, filename, param)
		}
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
		utils.LogInfo("%s [%d] %s", r.Method, statusCode, r.URL)
	})
}

func getTitle(filename string) string {
	return filepath.Base(filename)
}

func getMode(param *Param) string {
	if param.ForceDarkMode {
		return darkMode
	} else if param.ForceLightMode {
		return lightMode
	}

	isDark, err := dark.IsDarkMode()
	utils.LogDebug("Debug [auto-detected dark mode]: isDark=%v, err=%v", isDark, err)

	if err != nil {
		return defaultMode
	}

	if isDark {
		return darkMode
	}

	return lightMode
}

func isDarkMode(param *Param) bool {
	return getMode(param) == darkMode
}

func getPort(host string, port int) (int, error) {
	var err error

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		utils.LogInfo("Skipping port %d: %v", port, err)
		listener, err = net.Listen("tcp", host+":0")
	}

	_ = listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		panic("could not cast Addr to TCPAddr")
	}

	if err != nil {
		return addr.Port, fmt.Errorf("TCP listener error: %w", err)
	}

	return addr.Port, nil
}
