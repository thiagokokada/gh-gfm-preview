package server

import "net/http"

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
