package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/asherzog/thisor/internal/authenticator"
)

func (router Router) Login(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := router.db.SessionManager.NewSession()
		if err != nil {
			router.logger.Error("auth state failure", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}
		http.Redirect(w, r, auth.AuthCodeURL(state), http.StatusTemporaryRedirect)
	}
}

func (router Router) Callback(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state == "" {
			router.logger.Error("invalid state")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: "invalid state"})
			return
		}
		if _, err := router.db.SessionManager.GetSession(state); err != nil {
			router.logger.Error("invalid state", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}

		code := r.URL.Query().Get("code")
		// Exchange an authorization code for a token.
		token, err := auth.Exchange(r.Context(), code)
		if err != nil {
			router.logger.Error("code exchange", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}

		idToken, err := auth.VerifyIDToken(r.Context(), token)
		if err != nil {
			router.logger.Error("token validation", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			router.logger.Error("profile claims", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusInternalServerError, Msg: err.Error()})
			return
		}

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "jwt", Value: token.AccessToken, Expires: expiration}
		http.SetCookie(w, &cookie)
		ctx := auth.Set(r.Context(), profile)
		v := auth.Get(ctx)
		fmt.Printf("\n\n%v\n\n", v)
		// Redirect to logged in page.
		http.Redirect(w, r.WithContext(ctx), "/user", http.StatusTemporaryRedirect)
	}
}
