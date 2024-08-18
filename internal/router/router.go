package router

import (
	"log/slog"
	"net/http"

	"github.com/asherzog/thisor/internal/espn"
	sloghttp "github.com/samber/slog-http"
)

type Router struct {
	logger     *slog.Logger
	espnClient *espn.Client
}

func NewRouter(lg *slog.Logger) http.Handler {
	router := &Router{logger: lg}
	router.espnClient = espn.NewClient()
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", router.ping())
	mux.HandleFunc("/schedule", router.GetSchedule())
	mux.HandleFunc("/odds/{id}", router.GetOdds())

	handler := sloghttp.Recovery(mux)
	handler = sloghttp.New(lg)(handler)

	return handler
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
