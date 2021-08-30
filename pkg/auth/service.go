package auth

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	authorizationHeader            = "Authorization"
	authorizationHeaderValuePrefix = "Bearer "
	cookieName                     = "speakerbob-session"
	sessionTTL                     = 24 * time.Hour
	cleanupInterval                = 1 * time.Hour
)

type Service struct {
	TokenProvider *TokenProvider
	UserProvider  *UserProvider
	states        StateManager

	Providers []Provider
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	if !s.Enabled() {
		return
	}

	router.HandleFunc("/user/", s.getUser).Methods(http.MethodGet)
	router.HandleFunc("/user/", s.updateUser).Methods(http.MethodPatch)

	router.HandleFunc("/login/", s.providerRedirect).Methods(http.MethodGet)
	router.HandleFunc("/logout/", s.logout).Methods(http.MethodGet)

	router.HandleFunc("/providers/", s.listProviders).Methods(http.MethodGet)
	router.HandleFunc("/callback/", s.callback).Methods(http.MethodGet)

	router.HandleFunc("/token/", s.listToken).Methods(http.MethodGet)
	router.HandleFunc("/token/", s.createToken).Methods(http.MethodPost)
	router.HandleFunc("/token/{tokenId}/", s.deleteToken).Methods(http.MethodDelete)
}

func (s *Service) Run(ctx context.Context) {
	var err error
	var now time.Time
	var expiredTokens []*Token
	var ticker *time.Ticker

	if !s.Enabled() {
		return
	}

	ticker = time.NewTicker(cleanupInterval)

	logrus.Info("starting auth service worker")
	for {
		select {
		case <-ctx.Done():
			break
		case <-ticker.C:
			logrus.Debug("starting token cleanup")
			now = time.Now()

			expiredTokens = []*Token{}

			for _, token := range s.TokenProvider.List() {
				if !token.ExpiresAt.IsZero() && token.ExpiresAt.Before(now) {
					expiredTokens = append(expiredTokens, token)
				}
			}

			err = s.TokenProvider.Delete(expiredTokens...)
			if err != nil {
				logrus.Errorf("Failed to cleanup expired tokens: %s", err.Error())
			}
		}
	}
}

func (s *Service) Handler(h http.Handler) http.Handler {
	if !s.Enabled() {
		return h
	}

	return &Handler{h: h, tokenProvider: s.TokenProvider}
}

func (s *Service) Enabled() bool {
	return len(s.Providers) > 0
}

func (s *Service) callback(w http.ResponseWriter, r *http.Request) {
	var user *User
	var provider Provider

	providerName := s.states.getProviderName(r.URL.Query().Get("state"))
	for _, p := range s.Providers {
		if p.Name() == providerName {
			provider = p
			break
		}
	}

	// the state is invalid
	if provider == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// verify request
	principal, userEmail, err := provider.VerifyCallback(r)
	if err != nil {
		if _, ok := err.(AccessDenied); ok {
			http.Redirect(w, r, "/permission-denied/", http.StatusFound)
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// see if the user exists and is bound to this principal
	user = s.UserProvider.GetByPrincipals(principal)
	if user.Id == "" {
		// see if this user exists for this email
		user = s.UserProvider.GetByEmail(userEmail)
		if user.Id == "" {
			// if we find the user by their email, bind the principal
			user.Principals = append(user.Principals, principal)
			if err = s.UserProvider.Save(user); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			logrus.Debugf("Bound new principal [%s] to [%s]", principal, user.Id)
		}
	}

	// create new user
	if user.Id == "" {
		newUser := NewUser()
		newUser.Principals = append(newUser.Principals, principal)
		newUser.Email = user.Email
		if err = s.UserProvider.Save(&newUser); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		user = &newUser

		logrus.Debugf("Registered new user [%s] with email [%s]", user.Id, user.Email)
	}

	// generate a new token
	newToken := NewToken()
	newToken.Type = Session
	newToken.UserId = user.Id
	newToken.ExpiresAt = time.Now().Add(sessionTTL)
	if err = s.TokenProvider.Save(&newToken); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create our cookie
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    newToken.Token,
		Expires:  newToken.ExpiresAt,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

// User
func (s *Service) getUser(w http.ResponseWriter, r *http.Request) {
	token, valid := s.TokenProvider.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user := s.UserProvider.Get(token.UserId)

	if json.NewEncoder(w).Encode(user) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Service) updateUser(w http.ResponseWriter, r *http.Request) {
	var currentUser *User
	var user User
	var err error

	token, valid := s.TokenProvider.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	currentUser = s.UserProvider.Get(token.UserId)

	// decode user request
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// names cannot be set to empty
	if user.Name == "" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// write user changes
	currentUser.Name = user.Name

	err = s.UserProvider.Save(currentUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Token
func (s *Service) listToken(w http.ResponseWriter, r *http.Request) {
	token, valid := s.TokenProvider.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rval := make([]*Token, 0)
	for _, t := range s.TokenProvider.List() {
		if t.UserId == token.UserId {
			rval = append(rval, t)
		}
	}

	if json.NewEncoder(w).Encode(rval) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Service) createToken(w http.ResponseWriter, r *http.Request) {
	token, valid := s.TokenProvider.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	targetToken := Token{}
	err := json.NewDecoder(r.Body).Decode(&targetToken)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	newToken := NewToken()
	newToken.Type = Bearer
	newToken.UserId = token.UserId
	newToken.Name = targetToken.Name

	if s.TokenProvider.Save(&newToken) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if json.NewEncoder(w).Encode(newToken) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Service) deleteToken(w http.ResponseWriter, r *http.Request) {
	token, valid := s.TokenProvider.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	targetToken := s.TokenProvider.Get(mux.Vars(r)["tokenId"])

	if targetToken.UserId != token.Id {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if s.TokenProvider.Delete(targetToken) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// External Auth
func (s *Service) providerRedirect(w http.ResponseWriter, r *http.Request) {
	for _, provider := range s.Providers {
		if provider.Name() == r.URL.Query().Get("provider") {
			provider.LoginRedirect(w, r, s.states.NewState(provider))
			return
		}
	}

	w.WriteHeader(http.StatusNotAcceptable)

}

func (s *Service) listProviders(w http.ResponseWriter, _ *http.Request) {
	result := make([]string, 0)

	var provider Provider
	for _, provider = range s.Providers {
		result = append(result, provider.Name())
	}

	if json.NewEncoder(w).Encode(result) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Logout
func (s *Service) logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     cookieName,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)

	token := s.TokenProvider.FromRequest(r)

	if token != nil && s.TokenProvider.Delete(token) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
