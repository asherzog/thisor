package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/asherzog/thisor/internal/espn"
	"github.com/go-chi/chi"
)

func (web Web) Week(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		prof := session.Values["prof"]
		user, ok := prof.(authenticator.Profile)
		if !ok {
			web.lg.Warn("user not logged in")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		user["path"] = "week"

		url := fmt.Sprintf("http://localhost:8080/api/schedule/week/%s", chi.URLParam(r, "id"))

		req, err := http.NewRequestWithContext(r.Context(), "GET", url, nil)
		if err != nil {
			web.lg.Error("user request err")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		req.SetBasicAuth(os.Getenv("BASIC_USER"), os.Getenv("BASIC_PASS"))
		req.Header.Add("Content-Type", "application/json")
		req.Close = true
		resp, err := web.client.Do(req)
		if err != nil {
			web.lg.Error("user request err")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		var week espn.Schedule
		if err := json.Unmarshal(body, &week); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		user["schedule"] = week
		lid := r.URL.Query().Get("lid")
		user["lid"] = lid

		workDir, _ := os.Getwd()
		base := filepath.Join(workDir, "/web/template/header.html")
		u := filepath.Join(workDir, "/web/template/picks.html")
		tmpl, err := template.ParseFiles(u, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "schedule", user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
