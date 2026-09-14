package auth

import (
	"container/heap"
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
	wsTokenParameterName           = "token"
	cookieName                     = "speakerbob-session"
	sessionTTL                     = 24 * time.Hour
	wsTokenTTL                     = 1 * time.Minute
	cleanupChannelBufferSize       = 64
)

type Service struct {
	TokenProvider *TokenProvider
	UserProvider  *UserProvider
	states        StateManager

	Providers []Provider

	cleanupCh chan cleanupItem
}

type createTokenResponse struct {
	Token
	AccessToken string `json:"token"`
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	if !s.Enabled() {
		return
	}

	s.cleanupCh = make(chan cleanupItem, cleanupChannelBufferSize)

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
	if !s.Enabled() {
		return
	}

	pending := &cleanupQueue{}
	heap.Init(pending)

	logrus.Debug("seeding token cleanup queue")
	for _, token := range s.TokenProvider.List() {
		if !token.ExpiresAt.IsZero() {
			heap.Push(pending, cleanupItem{id: token.Id, expiresAt: token.ExpiresAt})
		}
	}

	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	resetTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		if pending.Len() == 0 {
			return
		}
		if d := time.Until((*pending)[0].expiresAt); d > 0 {
			timer.Reset(d)
		} else {
			timer.Reset(0)
		}
	}
	resetTimer()

	logrus.Info("starting auth service worker")
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-s.cleanupCh:
			heap.Push(pending, item)
			resetTimer()
		case <-timer.C:
			logrus.Debug("starting token cleanup")
			now := time.Now()

			var expiredTokens []*Token
			for pending.Len() > 0 && !(*pending)[0].expiresAt.After(now) {
				item := heap.Pop(pending).(cleanupItem)

				token := s.TokenProvider.Get(item.id)
				if token == nil || token.ExpiresAt.IsZero() || token.ExpiresAt.After(now) {
					continue
				}

				expiredTokens = append(expiredTokens, token)
			}

			if len(expiredTokens) > 0 {
				if err := s.TokenProvider.Delete(expiredTokens...); err != nil {
					logrus.Errorf("Failed to cleanup expired tokens: %s", err.Error())
				}
			}
			resetTimer()
		}
	}
}

// enqueueCleanup schedules a token for expiry-based deletion instead of
// waiting for the next full-list scan. Safe to call before Run starts
// (RegisterRoutes always initializes cleanupCh first); a full queue drops
// the item with a warning rather than blocking the caller.
func (s *Service) enqueueCleanup(item cleanupItem) {
	if s.cleanupCh == nil {
		return
	}

	select {
	case s.cleanupCh <- item:
	default:
		logrus.Warnf("cleanup queue full, token %q may not be cleaned up until next restart", item.id)
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
	logrus.Debug("[auth.callback] verifying callback")
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
	if user == nil {
		logrus.Debug("[auth.callback] user matching principal not found")
		// see if this user exists for this email
		user = s.UserProvider.GetByEmail(userEmail)
		if user != nil {
			// if we find the user by their email, bind the principal
			user.Principals = append(user.Principals, principal)
			if err = s.UserProvider.Save(user); err != nil {
				logrus.Errorf("[auth.callback] failed to update user: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			logrus.Debugf("Bound new principal [%s] to [%s]", principal, user.Id)
		}
	}

	// create new user
	if user == nil {
		logrus.Debug("[auth.callback] user does not exist")
		newUser := NewUser()
		newUser.Principals = append(newUser.Principals, principal)
		newUser.Email = userEmail
		if err = s.UserProvider.Save(&newUser); err != nil {
			logrus.Errorf("[auth.callback] failed to save new user: %v", err)
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
		logrus.Errorf("[auth.callback] failed to save new token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.enqueueCleanup(cleanupItem{id: newToken.Id, expiresAt: newToken.ExpiresAt})

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

	if err := json.NewEncoder(w).Encode(user.Preferences); err != nil {
		logrus.Errorf("[auth.getUserPreferences] failed to parse user preferences: %v", err)
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
		logrus.Errorf("[auth.updateUserPreferences] failed to save user preferences: %v", err)
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

	if err := json.NewEncoder(w).Encode(rval); err != nil {
		logrus.Errorf("[auth.listToken] failed to encode token list: %v", err)
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

	if err := s.TokenProvider.Save(&token); err != nil {
		logrus.Errorf("[auth.createWSToken] failed to save token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.enqueueCleanup(cleanupItem{id: token.Id, expiresAt: token.ExpiresAt})

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

	if err := s.TokenProvider.Save(&token); err != nil {
		logrus.Errorf("[auth.createToken] failed to save token: %v", err)
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

	if targetToken.UserId == token.UserId {
		if err := s.TokenProvider.Delete(targetToken); err != nil {
			logrus.Errorf("[auth.deleteToken] failed to delete token: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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

	_ = json.NewEncoder(w).Encode(result)
}

// Logout
func (s *Service) logout(w http.ResponseWriter, r *http.Request) {
	token := s.TokenProvider.FromRequest(r)

	if token != nil {
		if err := s.TokenProvider.Delete(token); err != nil {
			logrus.Errorf("[auth.logout] failed to delete token: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
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
