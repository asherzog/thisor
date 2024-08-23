package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/go-chi/chi"
)

type Web struct{}

func NewClient() *Web {
	return &Web{}
}

func (Web) Serve() chi.Router {
	r := chi.NewRouter()

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "/web"))
	FileServer(r, "/", filesDir)

	return r
}

func (Web) ServeStatic() chi.Router {
	r := chi.NewRouter()
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "/web/static"))
	FileServer(r, "/", filesDir)
	return r
}

func (Web) User(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.Get(r.Context())
		fmt.Printf("\n\n%v\n\n", user)
		workDir, _ := os.Getwd()
		fp := filepath.Join(workDir, "/web/template/user.html")
		tmpl, err := template.ParseFiles(fp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
