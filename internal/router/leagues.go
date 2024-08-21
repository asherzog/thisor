package router

import (
	"encoding/json"
	"net/http"

	"github.com/asherzog/thisor/internal/db"
	"github.com/go-chi/chi"
)

func (router Router) Leagues() chi.Router {
	r := chi.NewRouter()
	r.Get("/", router.GetAllLeagues())
	r.Post("/", router.CreateLeague())

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", router.GetLeague())                    // GET /leagues/{id}
		r.Put("/users", router.AddUserToLeague())         // PUT /leagues/{id}
		r.Delete("/users", router.DeleteUserFromLeague()) // DELETE /leagues/{id}
	})

	return r
}

func (router Router) GetLeague() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetLeague(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) AddUserToLeague() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user db.User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&user)
		if err != nil {
			router.logger.Error("invalid league request", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}

		res, err := router.db.AddUserToLeague(r.Context(), chi.URLParam(r, "id"), user.ID)
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) DeleteUserFromLeague() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user db.User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&user)
		if err != nil {
			router.logger.Error("invalid league request", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}

		res, err := router.db.DeleteUserFromLeague(r.Context(), chi.URLParam(r, "id"), user.ID)
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) GetAllLeagues() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetAllLeagues(r.Context())
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

// func (router Router) GetLeaguesForUser() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		res, err := router.db.GetLeaguesForUser(r.Context(), chi.URLParam(r, "id"))
// 		if err != nil {
// 			router.logger.Error("db error", "err", err.Error())
// 			w.WriteHeader(http.StatusBadRequest)
// 		}
// 		json.NewEncoder(w).Encode(res)
// 	}
// }

func (router Router) CreateLeague() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var league db.League
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&league)
		if err != nil {
			router.logger.Error("invalid league request", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}
		res, err := router.db.CreateLeague(r.Context(), league)
		if err != nil {
			router.logger.Error("db error", "err", err.Error(), "league", league)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}
