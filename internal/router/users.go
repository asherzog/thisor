package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asherzog/thisor/internal/db"
	"github.com/go-chi/chi"
)

func (router Router) Users() chi.Router {
	r := chi.NewRouter()
	r.Get("/", router.GetAllUsers())
	r.Post("/", router.PostUser())

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", router.GetUser())
	})

	return r
}

func (router Router) GetAllUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := router.db.GetAllUsers(r.Context())
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("\n\nGot ID: %s\n\n", chi.URLParam(r, "id"))
		res, err := router.db.GetUser(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (router Router) PostUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user db.User
		err := r.ParseForm()
		if err != nil {
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&user)
			if err != nil {
				router.logger.Error("invalid user request", "err", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(user)
				return
			}
		} else {
			user.ID = r.FormValue("id")
			user.Name = r.FormValue("name")
		}

		res, err := router.db.CreateUser(r.Context(), user)
		if err != nil {
			router.logger.Error("db error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(res)
	}
}
