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

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/asherzog/thisor/internal/espn"
	"github.com/go-chi/chi"
)

func (web *Web) Week(auth *authenticator.Authenticator) http.HandlerFunc {
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
		prof["path"] = "week"

		// get user info and picks
		uid, _ := url.PathUnescape(r.URL.Query().Get("uid"))
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

		week, err := web.getWeek(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			web.lg.Error("week request error", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		l := len(week.Games)
		if l == 0 {
			web.lg.Error("no games")
			http.Error(w, "no games in week", http.StatusInternalServerError)
		}
		sort.Slice(week.Games, func(i, j int) bool {
			return week.Games[i].Date < week.Games[j].Date
		})

		prof["selection"] = map[string]string{}
		prof["winScore"] = map[string]int{}
		prof["loseScore"] = map[string]int{}
		prof["resultWinner"] = map[string]string{}
		prof["resultWinScore"] = map[string]string{}
		prof["resultLoseScore"] = map[string]string{}
		prof["isWin"] = map[string]string{}
		for _, g := range week.Games {
			rw, _ := prof["resultWinner"].(map[string]string)
			rwscore, _ := prof["resultWinScore"].(map[string]string)
			rlscore, _ := prof["resultLoseScore"].(map[string]string)
			if g.Winner.ID != "" {
				rw[g.ID] = g.Winner.ID
				rwscore[g.ID] = g.WinScore
				rlscore[g.ID] = g.LoseScore
			} else {
				rw[g.ID] = "n/a"
			}
			prof["resultWinner"] = rw
			prof["resultWinScore"] = rwscore
			prof["resultLoseScore"] = rlscore
			for _, p := range user.Picks {
				if g.ID == p.GameID {
					prof["isLocked"] = p.IsLocked
					w, _ := prof["selection"].(map[string]string)
					w[p.GameID] = p.Selection.ID
					prof["selection"] = w
					if p.WinScore > 0 {
						ws, _ := prof["winScore"].(map[string]int)
						ls, _ := prof["loseScore"].(map[string]int)
						ws[p.GameID] = p.WinScore
						ls[p.GameID] = p.LoseScore
						prof["winScore"] = ws
						prof["loseScore"] = ls
					}
					isWin, _ := prof["isWin"].(map[string]string)
					isWin[p.GameID] = "n/a"
					if g.Winner.ID != "" {
						isWin[p.GameID] = "false"
						if p.Selection.ID == g.Winner.ID {
							isWin[p.GameID] = "true"
						}
					}
					prof["isWin"] = isWin
				}
			}
		}

		last := week.Games[len(week.Games)-1].ID
		prof["last"] = last
		prof["schedule"] = week
		lid := r.URL.Query().Get("lid")
		prof["lid"] = lid

		if len(web.weeks) == 0 {
			for i := 1; i < 19; i++ {
				web.weeks = append(web.weeks, i)
			}
		}
		prof["weeks"] = web.weeks

		prof["canSubmit"] = false
		loggedIn := prof["sub"].(string)
		if loggedIn == uid {
			prof["canSubmit"] = true
		}

		workDir, _ := os.Getwd()
		base := filepath.Join(workDir, "/web/template/header.html")
		league := filepath.Join(workDir, "/web/template/league.html")
		u := filepath.Join(workDir, "/web/template/picks.html")
		tmpl, err := template.ParseFiles(u, league, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "schedule", prof); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (web Web) getWeek(ctx context.Context, id string) (espn.Schedule, error) {
	var week espn.Schedule
	url := fmt.Sprintf("http://localhost:8080/api/schedule/week/%s", id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return week, err
	}
	req.SetBasicAuth(os.Getenv("BASIC_USER"), os.Getenv("BASIC_PASS"))
	req.Header.Add("Content-Type", "application/json")
	req.Close = true
	resp, err := web.client.Do(req)
	if err != nil {
		return week, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return week, err
	}

	if err := json.Unmarshal(body, &week); err != nil {
		return week, err
	}
	return week, nil
}
