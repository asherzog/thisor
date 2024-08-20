package router

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (router Router) handleScheduleRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch strings.ToLower(r.Method) {
		case "get":
			router.GetSchedule(w, r)
		case "put":
			router.PutSchedule(w, r)
		default:
			router.logger.Error("unsupported method", "method", r.Method, "request", r.Body)
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode("unsupported")
		}
	}
}

func (router Router) GetSchedule(w http.ResponseWriter, r *http.Request) {
	res, err := router.db.GetSchedule(r.Context(), "2024")
	if err != nil {
		router.logger.Error("db error", "err", err.Error())
		router.logger.Info("attempting to fetch from espn")
		res, err = router.espnClient.GetSchedule()
		if err != nil {
			router.logger.Error("espn error", "err", err.Error())
			// TODO: check for different types of errors
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	json.NewEncoder(w).Encode(res)
}

func (router Router) GetWeek() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := r.PathValue("id")
		// string to int
		id, err := strconv.Atoi(s)
		if err != nil {
			router.logger.Error("invalid week", "id", s)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("invalid week")
			return
		}
		res, err := router.db.GetWeek(r.Context(), id)
		if err != nil {
			router.logger.Error("invalid week", "id", s)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("invalid week")
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) PutSchedule(w http.ResponseWriter, r *http.Request) {
	res, err := router.espnClient.GetSchedule()
	if err != nil {
		router.logger.Error("espn error", "err", err.Error())
		// TODO: check for different types of errors
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	schedule, err := router.db.AddSchedule(r.Context(), res)
	if err != nil {
		router.logger.Error("db error", "err", err.Error())
	}
	json.NewEncoder(w).Encode(schedule)
}
