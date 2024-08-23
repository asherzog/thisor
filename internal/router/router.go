package router

import (
	"context"
	"log/slog"
	"net/http"

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
	router.web = web.NewClient()

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Get("/ping", router.ping())
	r.Mount("/home", router.web.Serve())
	r.Mount("/login", router.Login(router.auth))
	r.Mount("/callback", router.Callback(router.auth))
	r.Mount("/user", router.web.User(router.auth))
	r.Mount("/picks", router.Picks())
	r.Mount("/schedule", router.Schedule())
	r.Mount("/odds", router.Odds())
	r.Mount("/users", router.requireUser(router.Users()))
	r.Mount("/leagues", router.Leagues())

	handler := sloghttp.Recovery(r)
	handler = sloghttp.New(lg)(handler)

	return handler, nil
}

func (Router) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ping" || r.Method != http.MethodGet {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}
}

func (router Router) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := router.auth.Get(r.Context())
		if user == nil {
			// No user so redirect to login
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		ctx := router.auth.Set(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
