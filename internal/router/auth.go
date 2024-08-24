package router

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/asherzog/thisor/internal/authenticator"
)

func (router Router) Login(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		state, err := generateRandomState()
		if err != nil {
			router.logger.Error("auth state failure", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: err.Error()})
			return
		}

		// one time use
		session.AddFlash(state)
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, auth.AuthCodeURL(state), http.StatusTemporaryRedirect)
	}
}

func (router Router) Callback(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		flashes := session.Flashes()
		if len(flashes) == 0 {
			router.logger.Error("invalid state")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: "invalid state"})
			return
		}

		state := r.URL.Query().Get("state")
		if flashes[0] != state {
			router.logger.Error("invalid state")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusBadRequest, Msg: "invalid state"})
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

		profile := auth.Profile
		if err := idToken.Claims(&profile); err != nil {
			router.logger.Error("profile claims", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorReturn{Status: http.StatusInternalServerError, Msg: err.Error()})
			return
		}

		session.Values["jwt"] = token.AccessToken
		session.Values["prof"] = profile

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to logged in page.
		http.Redirect(w, r, "/user", http.StatusTemporaryRedirect)
	}
}

func (router Router) Logout(auth *authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Options.MaxAge = -1
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		returnTo, err := url.Parse(scheme + "://" + r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		parameters := url.Values{}
		parameters.Add("returnTo", returnTo.String())
		parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
		logoutUrl.RawQuery = parameters.Encode()

		http.Redirect(w, r, logoutUrl.String(), http.StatusTemporaryRedirect)
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.StdEncoding.EncodeToString(b)
	return state, nil
}
