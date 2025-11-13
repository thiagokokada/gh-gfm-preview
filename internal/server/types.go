package server

import (
	"net/http"
	"regexp"
)

type TemplateParam struct {
	Title  string
	Body   string
	Host   string
	Reload bool
	Mode   string
}

type Param struct {
	Filename       string
	MarkdownMode   bool
	Reload         bool
	ForceLightMode bool
	ForceDarkMode  bool
	AutoOpen       bool
	UseStdin       bool
	StdinContent   string
}

type Server struct {
	Host string
	Port int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

type capturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (c *capturingResponseWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	// Don't call the underlying WriteHeader yet - we'll handle it later
}

func (c *capturingResponseWriter) Write(b []byte) (int, error) {
	c.body = append(c.body, b...)
	return len(b), nil
}

func (c *capturingResponseWriter) ExtractDirectoryListingBody() string {
	html := string(c.body)

	// Remove <!doctype html> and <meta> tags that FileServer adds
	html = regexp.MustCompile(`(?i)<!doctype[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)<meta[^>]*>`).ReplaceAllString(html, "")

	// Return the cleaned HTML
	return html
}

type mdResponseJSON struct {
	HTML  string `json:"html"`
	Title string `json:"title"`
}
