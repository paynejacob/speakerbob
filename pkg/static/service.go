package static

import (
	"github.com/gorilla/mux"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

//go:generate go run generator.go

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s Service) RegisterRoutes(parent *mux.Router, prefix string) {
	parent.PathPrefix(prefix).Handler(s)
}

func (s Service) Run() {
	return
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if r.URL.Path == "" || r.URL.Path == "/" {
		path = "index.html"
	}

	value, ok := StaticMap[path]
	if !ok {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	w.Header().Set("Content-Encoding", "gzip")
	setContentType(w, path)
	setCacheHeader(w, path)
	_, _ = w.Write(value)
}

func setContentType(w http.ResponseWriter, path string) {
	extension := filepath.Ext(path)
	mimeType := mime.TypeByExtension(extension)

	if mimeType == "" {
		return
	}

	w.Header().Set("Content-Type", mimeType)
}

func setCacheHeader(w http.ResponseWriter, path string) {
	extension := filepath.Ext(path)

	if extension == ".js" {
		w.Header().Set("Cache-Control", "max-age=2592000s") // 30d
	} else if extension == ".css" {
		w.Header().Set("Cache-Control", "max-age=2592000s") // 30d
	} else if extension == ".ico" {
		w.Header().Set("Cache-Control", "max-age=2592000s") // 30d
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
}
