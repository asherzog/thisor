package web

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/go-chi/chi"
)

type Web struct {
	client     *http.Client
	lg         *slog.Logger
	isLoggedIn bool
	weeks      []int
}

func NewClient(lg *slog.Logger) *Web {
	client := http.Client{
		Timeout: time.Second * 10,
	}
	return &Web{lg: lg, client: &client}
}

func (web *Web) Serve(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		val := session.Values["prof"]
		user, ok := val.(authenticator.Profile)
		if !ok {
			web.isLoggedIn = false
			user = make(map[string]interface{})
		}
		user["path"] = "home"

		workDir, _ := os.Getwd()
		base := filepath.Join(workDir, "/web/template/header.html")
		h := filepath.Join(workDir, "/web/template/home.html")
		tmpl, err := template.ParseFiles(h, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "home", user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (Web) ServeStatic() chi.Router {
	r := chi.NewRouter()
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "/web/static"))
	FileServer(r, "/", filesDir)
	return r
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
