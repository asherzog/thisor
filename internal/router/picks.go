package router

import (
	"encoding/json"
	"net/http"

	"github.com/asherzog/thisor/internal/db"
	"github.com/go-chi/chi"
)

func (router Router) Picks() chi.Router {
	r := chi.NewRouter()
	r.Get("/", router.GetAllPicks())
	r.Post("/", router.PostPick())
	r.Put("/", router.PostPick())
	r.Post("/list", router.PostPickList())
	r.Put("/list", router.PostPickList())

	r.Route("/users/{id}", func(r chi.Router) {
		r.Get("/", router.GetPicksForUser()) // GET /picks/user/{id}
	})

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", router.GetPick()) // GET /picks/{id}
	})

	return r
}

func (router Router) GetPick() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetPick(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) GetAllPicks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetAllPicks(r.Context())
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) GetPicksForUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetPicksForUser(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) PostPickList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var picks db.PickList
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&picks)
		if err != nil {
			router.logger.Error("invalid pick request", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}
		res, err := router.db.PostPickList(r.Context(), picks)
		if err != nil {
			router.logger.Error("db error", "err", err.Error(), "pick", picks)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) PostPick() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pick db.Pick
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&pick)
		if err != nil {
			router.logger.Error("invalid pick request", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}
		res, err := router.db.CreatePick(r.Context(), pick)
		if err != nil {
			router.logger.Error("db error", "err", err.Error(), "pick", pick)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}
