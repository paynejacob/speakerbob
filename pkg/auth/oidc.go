package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	oidcProviderName = "oidc"
	oidcStateTTL     = 5 * time.Minute
)

type OIDCConfig struct {
	Enabled             bool     `yaml:"enabled"`
	Issuer              string   `yaml:"issuer"`
	ClientId            string   `yaml:"client_id"`
	ClientSecret        string   `yaml:"client_secret"`
	RedirectURL         string   `yaml:"redirect_url"`
	Scopes              []string `yaml:"scopes"`
	AllowedEmailDomains []string `yaml:"allowed_email_domains"`
}

type oidcState struct {
	nonce    string
	verifier string
	expires  time.Time
}

type oidcClient struct {
	oauth    oauth2.Config
	verifier *oidc.IDTokenVerifier
}

type oidcFlow struct {
	mu     sync.Mutex
	states map[string]oidcState
	client *oidcClient
}

func (f *oidcFlow) newState() (state string, entry oidcState, err error) {
	if state, err = randomString(); err != nil {
		return
	}
	if entry.nonce, err = randomString(); err != nil {
		return
	}
	entry.verifier = oauth2.GenerateVerifier()
	entry.expires = time.Now().Add(oidcStateTTL)

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.states == nil {
		f.states = make(map[string]oidcState)
	}
	for k, v := range f.states {
		if time.Now().After(v.expires) {
			delete(f.states, k)
		}
	}
	f.states[state] = entry

	return
}

// consume is one-shot: a state can only complete a single callback.
func (f *oidcFlow) consume(state string) (oidcState, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry, ok := f.states[state]
	delete(f.states, state)

	return entry, ok && time.Now().Before(entry.expires)
}

func randomString() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (s *Service) oidcClient(ctx context.Context) (*oidcClient, error) {
	s.oidc.mu.Lock()
	defer s.oidc.mu.Unlock()

	if s.oidc.client != nil {
		return s.oidc.client, nil
	}

	provider, err := oidc.NewProvider(ctx, s.OIDC.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to discover oidc issuer %q: %w", s.OIDC.Issuer, err)
	}

	scopes := s.OIDC.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "email", "profile"}
	}

	s.oidc.client = &oidcClient{
		oauth: oauth2.Config{
			ClientID:     s.OIDC.ClientId,
			ClientSecret: s.OIDC.ClientSecret,
			RedirectURL:  s.OIDC.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       scopes,
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: s.OIDC.ClientId}),
	}

	return s.oidc.client, nil
}

func (s *Service) oidcLogin(w http.ResponseWriter, r *http.Request) {
	client, err := s.oidcClient(r.Context())
	if err != nil {
		logrus.Errorf("[auth.oidcLogin] %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	state, entry, err := s.oidc.newState()
	if err != nil {
		logrus.Errorf("[auth.oidcLogin] failed to generate state: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url := client.oauth.AuthCodeURL(state, oidc.Nonce(entry.nonce), oauth2.S256ChallengeOption(entry.verifier))
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Service) oidcCallback(w http.ResponseWriter, r *http.Request) {
	entry, ok := s.oidc.consume(r.URL.Query().Get("state"))
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	client, err := s.oidcClient(r.Context())
	if err != nil {
		logrus.Errorf("[auth.oidcCallback] %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	oauthToken, err := client.oauth.Exchange(r.Context(), r.URL.Query().Get("code"), oauth2.VerifierOption(entry.verifier))
	if err != nil {
		logrus.Errorf("[auth.oidcCallback] code exchange failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rawIDToken, _ := oauthToken.Extra("id_token").(string)
	idToken, err := client.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		logrus.Errorf("[auth.oidcCallback] id token verification failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if idToken.Nonce != entry.nonce {
		logrus.Error("[auth.oidcCallback] id token nonce mismatch")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err = idToken.Claims(&claims); err != nil {
		logrus.Errorf("[auth.oidcCallback] failed to parse claims: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if claims.Email == "" || !claims.EmailVerified || !s.oidcEmailAllowed(claims.Email) {
		http.Redirect(w, r, "/permission-denied/", http.StatusFound)
		return
	}

	s.loginUser(w, r, NewPrincipal(oidcProviderName, idToken.Subject), claims.Email)
}

func (s *Service) oidcEmailAllowed(email string) bool {
	if len(s.OIDC.AllowedEmailDomains) == 0 {
		return true
	}

	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}

	for _, domain := range s.OIDC.AllowedEmailDomains {
		if strings.EqualFold(email[at+1:], domain) {
			return true
		}
	}

	return false
}
