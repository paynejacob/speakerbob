package auth

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/service"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	authorizationHeader            = "Authorization"
	authorizationHeaderValuePrefix = "Bearer "
	wsTokenParameterName           = "auth"
	cookieName                     = "speakerbob-session"
	sessionTTL                     = 24 * time.Hour
	wsTokenTTL                     = 1 * time.Minute
	cleanupInterval                = 1 * time.Hour
)

type Service struct {
	TokenProvider *TokenProvider
	UserProvider  *UserProvider
	states        StateManager

	Providers []Provider
}

type createTokenResponse struct {
	Token
	AccessToken string `json:"token"`
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	if !s.Enabled() {
		return
	}

	router.HandleFunc("/user/preferences/", s.getUserPreferences).Methods(http.MethodGet)
	router.HandleFunc("/user/preferences/", s.updateUserPreferences).Methods(http.MethodPatch)

	router.HandleFunc("/login/", s.providerRedirect).Methods(http.MethodGet)
	router.HandleFunc("/logout/", s.logout).Methods(http.MethodGet)

	router.HandleFunc("/providers/", s.listProviders).Methods(http.MethodGet)
	router.HandleFunc("/callback/", s.callback).Methods(http.MethodGet)

	router.HandleFunc("/tokens/", s.listToken).Methods(http.MethodGet)
	router.HandleFunc("/tokens/ws/", s.createWSToken).Methods(http.MethodGet)
	router.HandleFunc("/tokens/", s.createToken).Methods(http.MethodPost)
	router.HandleFunc("/tokens/{tokenId}/", s.deleteToken).Methods(http.MethodDelete)
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

	return &Handler{h: h, AuthService: s}
}

func (s *Service) Enabled() bool {
	return len(s.Providers) > 0
}

func (s *Service) VerifyRequest(r *http.Request) (*Token, bool) {
	return s.verifyRequest(r, Bearer, Session)
}

func (s *Service) VerifyWebsocket(r *http.Request) (*Token, bool) {
	return s.verifyRequest(r, Bearer, Session, Websocket)
}

// Callback
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
	if user != nil {
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
	if user == nil {
		newUser := NewUser()
		newUser.Principals = append(newUser.Principals, principal)
		newUser.Email = userEmail
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
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

// User
func (s *Service) getUserPreferences(w http.ResponseWriter, r *http.Request) {
	token, valid := s.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user := s.UserProvider.Get(token.UserId)

	if json.NewEncoder(w).Encode(user.Preferences) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Service) updateUserPreferences(w http.ResponseWriter, r *http.Request) {
	var err error
	var user *User
	var requestPreferences map[string]string

	token, valid := s.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user = s.UserProvider.Get(token.UserId)

	// decode user request
	err = json.NewDecoder(r.Body).Decode(&requestPreferences)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// if the user has no preferences use the request preferences, otherwise merge the request preferences left
	if user.Preferences == nil {
		user.Preferences = requestPreferences
	} else {
		for k, v := range requestPreferences {
			// empty keys are deleted
			if v == "" {
				delete(user.Preferences, k)
				continue
			}

			user.Preferences[k] = v
		}
	}

	err = s.UserProvider.Save(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Token
func (s *Service) listToken(w http.ResponseWriter, r *http.Request) {
	token, valid := s.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rval := make([]*Token, 0)
	for _, t := range s.TokenProvider.List() {
		if t.UserId == token.UserId && t.Type == Bearer {
			rval = append(rval, t)
		}
	}

	if json.NewEncoder(w).Encode(rval) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Service) createWSToken(w http.ResponseWriter, r *http.Request) {
	var token Token
	var userId string

	if t, valid := s.VerifyRequest(r); !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		userId = t.UserId
	}

	token = NewToken()
	token.Type = Websocket
	token.UserId = userId
	token.ExpiresAt = time.Now().Add(wsTokenTTL)

	if s.TokenProvider.Save(&token) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := createTokenResponse{token, token.Token}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Service) createToken(w http.ResponseWriter, r *http.Request) {
	var err error

	var token Token
	var requestToken Token
	var userId string

	if t, valid := s.VerifyRequest(r); !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		userId = t.UserId
	}

	err = json.NewDecoder(r.Body).Decode(&requestToken)
	if err != nil {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("unable to parse request"))
		return
	}

	token = NewToken()
	token.Type = Bearer
	token.UserId = userId
	token.Name = requestToken.Name

	if s.TokenProvider.Save(&token) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := createTokenResponse{token, token.Token}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Service) deleteToken(w http.ResponseWriter, r *http.Request) {
	token, valid := s.VerifyRequest(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	targetToken := s.TokenProvider.Get(mux.Vars(r)["tokenId"])

	if targetToken.UserId == token.UserId && s.TokenProvider.Delete(targetToken) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	token := s.TokenProvider.FromRequest(r)

	if token != nil && s.TokenProvider.Delete(token) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Secure:   true,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusNoContent)
}

// Utils
func (s *Service) verifyRequest(r *http.Request, allowedTypes ...TokenType) (*Token, bool) {
	if !s.Enabled() {
		return nil, true
	}

	token := s.TokenProvider.FromRequest(r)

	if token == nil {
		return nil, false
	}

	var allowed bool
	for i := 0; i < len(allowedTypes); i++ {
		allowed = token.Type == allowedTypes[i]

		if allowed {
			break
		}
	}

	return token, (token.ExpiresAt.IsZero() || time.Now().Before(token.ExpiresAt)) && allowed
}
