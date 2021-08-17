package static

import (
	"context"
	"embed"
	"github.com/gorilla/mux"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

const assetPath = "assets"
const indexPath = "assets/index.html"

//go:embed assets
var assetsFS embed.FS

type FS struct {
	fs.FS
}

func (f FS) Open(name string) (fs.File, error) {
	name = filepath.Join(assetPath, name)
	r, err := f.FS.Open(name)
	if os.IsNotExist(err) {
		return f.FS.Open(indexPath)
	}

	return r, err
}

type Service struct{}

func (s Service) RegisterRoutes(router *mux.Router) {
	router.PathPrefix("/").Handler(s)
}

func (s Service) Run(context.Context) {
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCacheHeader(w, r)
	http.FileServer(http.FS(FS{assetsFS})).ServeHTTP(w, r)
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
