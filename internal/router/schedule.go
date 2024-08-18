package router

import (
	"encoding/json"
	"net/http"
)

func (router Router) GetSchedule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.espnClient.GetSchedule()
		if err != nil {
			router.logger.Error("espn error", "err", err.Error())
			// TODO: check for different types of errors
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}
