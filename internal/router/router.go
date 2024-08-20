package router

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/asherzog/thisor/internal/db"
	"github.com/asherzog/thisor/internal/espn"
	sloghttp "github.com/samber/slog-http"
)

type Router struct {
	logger     *slog.Logger
	espnClient *espn.Client
	db         *db.DB
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

	mux := http.NewServeMux()

	mux.HandleFunc("/ping", router.ping())
	mux.HandleFunc("/schedule/week/{id}", router.GetWeek())
	mux.HandleFunc("/schedule", router.handleScheduleRequest())
	mux.HandleFunc("/odds/{id}", router.GetOdds())
	mux.HandleFunc("/games/{id}", router.GetGame())

	handler := sloghttp.Recovery(mux)
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
		w.Write([]byte("ok"))
	}
}
