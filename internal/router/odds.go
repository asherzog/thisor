package router

import (
	"encoding/json"
	"net/http"
)

func (router Router) GetOdds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.espnClient.GetGameOdds(r.PathValue("id"))
		if err != nil {
			router.logger.Error("espn error", "err", err.Error())
			// TODO: check for different types of errors
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}
