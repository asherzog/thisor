package web

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/asherzog/thisor/internal/espn"
)

func (web Web) Schedule(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		val := session.Values["prof"]
		prof, ok := val.(authenticator.Profile)
		if !ok {
			web.lg.Warn("user not logged in")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// So we know which page we are on in templates
		prof["path"] = "schedule"

		url := "http://localhost:8080/api/schedule"

		req, err := http.NewRequestWithContext(r.Context(), "GET", url, nil)
		if err != nil {
			web.lg.Error("schedule request err")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		req.SetBasicAuth(os.Getenv("BASIC_USER"), os.Getenv("BASIC_PASS"))
		req.Header.Add("Content-Type", "application/json")
		req.Close = true
		resp, err := web.client.Do(req)
		if err != nil {
			web.lg.Error("schedule request err")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		var schedule espn.Schedule
		if err := json.Unmarshal(body, &schedule); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		prof["schedule"] = schedule

		workDir, _ := os.Getwd()
		base := filepath.Join(workDir, "/web/template/header.html")
		u := filepath.Join(workDir, "/web/template/picks.html")
		tmpl, err := template.ParseFiles(u, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "schedule", prof); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
