package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/asherzog/thisor/internal/db"
	"github.com/go-chi/chi"
)

func (web *Web) User(auth *authenticator.Authenticator) http.HandlerFunc {
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
		prof["path"] = "user"

		uid, _ := url.PathUnescape(chi.URLParam(r, "id"))
		prof["withUser"] = true
		if uid == "" {
			uid = prof["sub"].(string)
			prof["withUser"] = false
		}
		prof["uid"] = uid
		user, err := web.getUser(r.Context(), uid, prof["sub"].(string))
		if err != nil {
			web.lg.Error("user request error", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		prof["user"] = user

		if len(web.weeks) == 0 {
			for i := 1; i < 19; i++ {
				web.weeks = append(web.weeks, i)
			}
		}
		prof["weeks"] = web.weeks

		prof["withLeague"] = false
		lid := r.URL.Query().Get("lid")
		if lid != "" {
			prof["withLeague"] = true
			prof["lid"] = lid
		}

		workDir, _ := os.Getwd()
		base := filepath.Join(workDir, "/web/template/header.html")
		league := filepath.Join(workDir, "/web/template/league.html")
		u := filepath.Join(workDir, "/web/template/user.html")
		tmpl, err := template.ParseFiles(u, league, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "user", prof); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (web *Web) UserCreate(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		val := session.Values["prof"]
		user, ok := val.(authenticator.Profile)
		if !ok {
			web.lg.Warn("user not logged in")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		user["path"] = "user"

		workDir, _ := os.Getwd()
		base := filepath.Join(workDir, "/web/template/header.html")
		u := filepath.Join(workDir, "/web/template/newUser.html")
		tmpl, err := template.ParseFiles(u, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "newuser", user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (web Web) getUser(ctx context.Context, id, sub string) (db.User, error) {
	var user db.User
	url := fmt.Sprintf("http://localhost:8080/api/users/%s", id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return user, err
	}
	req.SetBasicAuth(os.Getenv("BASIC_USER"), os.Getenv("BASIC_PASS"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("sub", sub)
	req.Close = true
	resp, err := web.client.Do(req)
	if err != nil {
		return user, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return user, err
	}

	if err := json.Unmarshal(body, &user); err != nil {
		return user, err
	}
	return user, nil
}
