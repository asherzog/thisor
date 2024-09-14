package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func (router Router) Schedule() chi.Router {
	r := chi.NewRouter()
	r.Get("/", router.GetSchedule())
	r.Post("/", router.PostSchedule())
	r.Put("/", router.PostSchedule())

	r.Route("/week/{id}", func(r chi.Router) {
		r.Get("/", router.GetWeek()) // GET /schedule/weeks/{id}
	})

	r.Route("/games/{id}", func(r chi.Router) {
		r.Get("/", router.GetGame()) // GET /schedule/games/{id}
	})

	return r
}

func (router Router) GetSchedule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.espnClient.GetSchedule()
		if err != nil {
			router.logger.Error("espn error", "err", err.Error())
			// TODO: check for different types of errors
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err.Error())
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) GetWeek() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "id")
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

func (router Router) PostSchedule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

func (router Router) GetGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetGame(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			// TODO: check for different types of errors
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}
