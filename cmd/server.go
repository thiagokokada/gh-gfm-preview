package cmd

import (
	"embed"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
)

type TemplateParam struct {
	Title  string
	Body   string
	Host   string
	Reload bool
	Mode   string
}

type Server struct {
	host string
	port int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

//go:embed template.html
var htmlTemplate string

//go:embed static/*
var staticDir embed.FS

const defaultPort = 3333
const isDarkModeDefault = true
const darkMode = "dark"
const lightMode = "light"

func (server *Server) Serve(param *Param) error {
	host := server.host
	port := defaultPort
	if server.port > 0 {
		port = server.port
	}

	filename, err := targetFile(param.filename)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)

	r := http.NewServeMux()
	r.Handle("/", wrapHandler(handler(filename, param, http.FileServer(http.Dir(dir)))))
	r.Handle("/static/", wrapHandler(handler(filename, param, http.FileServer(http.FS(staticDir)))))
	r.Handle("/__/md", wrapHandler(mdHandler(filename, param)))

	watcher, err := createWatcher(dir)
	if err != nil {
		return err
	}
	r.Handle("/ws", wsHandler(watcher))

	port, err = getPort(host, port)
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%d", host, port)

	logInfo("Accepting connections at http://%s/\n", address)

	if param.autoOpen {
		logInfo("Open http://%s/ on your browser\n", address)
		go func() {
			err := openBrowser(fmt.Sprintf("http://%s/", address))
			if err != nil {
				logInfo("Error while opening browser: %s\n", err)
			}
		}()
	}

	err = http.ListenAndServe(address, r)
	if err != nil {
		return err
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

		tmpl := template.Must(template.New("HTML Template").Parse(htmlTemplate))

		markdown, err := slurp(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		html, err := toHTML(markdown, param)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		param := TemplateParam{
			Title:  getTitle(filename),
			Body:   html,
			Host:   r.Host,
			Reload: param.reload,
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

	markdown, err := slurp(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	html, err := toHTML(markdown, param)
	if err != nil {
		http.Error(w, err.Error(), 500)
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

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func wrapHandler(wrappedHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, r)

		statusCode := lrw.statusCode
		logInfo("%s [%d] %s", r.Method, statusCode, r.URL)
	})
}

func getTitle(filename string) string {
	return filepath.Base(filename)
}

func getMode(param *Param) string {
	if param.forceDarkMode {
		return darkMode
	} else if param.forceLightMode {
		return lightMode
	} else if isDarkMode() {
		return darkMode
	} else {
		return lightMode
	}
}

func getPort(host string, port int) (int, error) {
	var err error
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		logInfo("%s", err.Error())
		listener, err = net.Listen("tcp", fmt.Sprintf("%s:0", host))
	}
	port = listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, err
}
