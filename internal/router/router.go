package router

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/asherzog/thisor/internal/authenticator"
	"github.com/asherzog/thisor/internal/db"
	"github.com/asherzog/thisor/internal/espn"
	"github.com/asherzog/thisor/web"
	"github.com/go-chi/chi"
	sloghttp "github.com/samber/slog-http"
)

type Router struct {
	logger     *slog.Logger
	espnClient *espn.Client
	db         *db.DB
	web        *web.Web
	auth       *authenticator.Authenticator
}

type ErrorReturn struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

func NewRouter(lg *slog.Logger) (http.Handler, error) {
	ctx := context.Background()
	dbClient, err := db.NewClient(ctx, lg)
	if err != nil {
		return nil, err
	}

	auth, err := authenticator.New()
	if err != nil {
		return nil, err
	}

	router := &Router{logger: lg, auth: auth}
	router.espnClient = espn.NewClient()
	router.db = dbClient
	router.web = web.NewClient(lg)

	r := chi.NewRouter()

	r.Mount("/", router.web.Serve(router.auth))
	r.Mount("/static", router.web.ServeStatic())
	r.Mount("/login", router.Login(router.auth))
	r.Mount("/callback", router.Callback(router.auth))
	r.Mount("/user", router.web.User(router.auth))
	r.Mount("/user/{id}", router.web.User(router.auth))
	r.Mount("/create-user", router.web.UserCreate(router.auth))
	r.Mount("/join", router.web.AddUserToLeague(router.auth))
	r.Mount("/schedule", router.web.Schedule(router.auth))
	r.Mount("/league/{id}", router.web.League(router.auth))
	r.Mount("/week/{id}", router.web.Week(router.auth))
	r.Mount("/logout", router.Logout(router.auth))
	r.Mount("/api", router.API())

	handler := sloghttp.Recovery(r)
	handler = sloghttp.New(lg)(handler)

	return handler, nil
}

func (Router) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}
}

// API routes
func (router Router) API() chi.Router {
	r := chi.NewRouter()

	r.Use(router.authenticatedRoute)

	r.Get("/ping", router.ping())
	r.Mount("/picks", router.Picks())
	r.Mount("/schedule", router.Schedule())
	r.Mount("/odds", router.Odds())
	r.Mount("/users", router.Users())
	r.Mount("/leagues", router.Leagues())
	return r
}

func (router Router) authenticatedRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/ping" {
			next.ServeHTTP(w, r)
			return
		}
		session, err := router.auth.Store.Get(r, "jwt")
		if err != nil {
			router.logger.Error("bad session", "error", err.Error())
			http.SetCookie(w, &http.Cookie{Name: "jwt", MaxAge: -1, Path: "/"})
			return
		}
		// is not authed browser client, check basic auth
		if session.IsNew {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusUnauthorized, Msg: "unauthorized"})
				return
			}

			if !isAuthorised(username, password) {
				w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusUnauthorized, Msg: "unauthorized"})
				return
			}
		}

		r = r.WithContext(context.WithValue(r.Context(), "session", session))
		next.ServeHTTP(w, r)
	})
}

func isAuthorised(user, pass string) bool {
	u := os.Getenv("BASIC_USER")
	p := os.Getenv("BASIC_PASS")

	return u == user && p == pass
}
