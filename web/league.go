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
	"sort"
	"strconv"

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/asherzog/thisor/internal/db"
	"github.com/go-chi/chi"
)

func (web *Web) League(auth *authenticator.Authenticator) http.HandlerFunc {
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
		prof["path"] = "league"

		// get user info and picks
		uid, _ := url.PathUnescape(r.URL.Query().Get("uid"))
		prof["withUser"] = true
		if uid == "" {
			uid = prof["sub"].(string)
			prof["withUser"] = false
		}
		user, err := web.getUser(r.Context(), uid, prof["sub"].(string))
		if err != nil {
			web.lg.Error("user request error", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		prof["user"] = user

		league, err := web.getLeague(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			web.lg.Error("league request error", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		prof["league"] = league

		var weeks []int
		for w := range league.Weeks {
			i, err := strconv.Atoi(w)
			if err != nil {
				return
			}
			weeks = append(weeks, i)
		}
		sort.Ints(weeks)
		prof["weeks"] = weeks
		prof["lid"] = league.ID
		web.weeks = weeks

		workDir, _ := os.Getwd()
		base := filepath.Join(workDir, "/web/template/header.html")
		u := filepath.Join(workDir, "/web/template/league.html")
		tmpl, err := template.ParseFiles(u, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "league", prof); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (web *Web) AddUserToLeague(auth *authenticator.Authenticator) http.HandlerFunc {
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
		u := filepath.Join(workDir, "/web/template/leagueJoin.html")
		tmpl, err := template.ParseFiles(u, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "leagueJoin", user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (web Web) getLeague(ctx context.Context, id string) (db.League, error) {
	var league db.League
	url := fmt.Sprintf("http://localhost:8080/api/leagues/%s", id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return league, err
	}
	req.SetBasicAuth(os.Getenv("BASIC_USER"), os.Getenv("BASIC_PASS"))
	req.Header.Add("Content-Type", "application/json")
	req.Close = true
	resp, err := web.client.Do(req)
	if err != nil {
		return league, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return league, err
	}

	if err := json.Unmarshal(body, &league); err != nil {
		return league, err
	}
	return league, nil
}
