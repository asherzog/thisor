package router

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/asherzog/thisor/internal/db"
	"github.com/asherzog/thisor/internal/espn"
	"github.com/go-chi/chi"
	sloghttp "github.com/samber/slog-http"
)

type Router struct {
	logger     *slog.Logger
	espnClient *espn.Client
	db         *db.DB
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

	router := &Router{logger: lg}
	router.espnClient = espn.NewClient()
	router.db = dbClient

	// mux := http.NewServeMux()
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Get("/ping", router.ping())
	r.Mount("/picks", router.Picks())
	r.Mount("/schedule", router.Schedule())
	r.Mount("/odds", router.Odds())
	r.Mount("/users", router.Users())
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
