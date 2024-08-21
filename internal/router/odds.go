package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

func (router Router) Odds() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", router.GetOdds())
	return r
}

func (router Router) GetOdds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.espnClient.GetGameOdds(chi.URLParam(r, "id"))
		if err != nil {
			router.logger.Error("espn error", "err", err.Error())
			// TODO: check for different types of errors
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}
