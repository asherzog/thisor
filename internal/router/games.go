package router

import (
	"encoding/json"
	"net/http"
)

func (router Router) GetGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetGame(r.Context(), r.PathValue("id"))
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			// TODO: check for different types of errors
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}
