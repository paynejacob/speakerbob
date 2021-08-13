package static

import (
	"embed"
	"github.com/gorilla/mux"
	"io/fs"
	"net/http"
	"path/filepath"
)

//go:embed assets
var assetsFS embed.FS

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
	rootFs, _ := fs.Sub(assetsFS, "assets")

	setCacheHeader(w, r)
	http.FileServer(http.FS(rootFs)).ServeHTTP(w, r)
}

func setCacheHeader(w http.ResponseWriter, r *http.Request) {
	extension := filepath.Ext(r.URL.Path)

	switch extension {
	case ".woff2":
		fallthrough
	case ".css":
		fallthrough
	case ".ico":
		fallthrough
	case ".js":
		w.Header().Set("Cache-Control", "max-age=31536000") // 1y
		break
	default:
		w.Header().Set("Cache-Control", "no-cache")
	}
}
