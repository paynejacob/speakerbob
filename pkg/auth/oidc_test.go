package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gorilla/mux"
	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeIdP struct {
	server  *httptest.Server
	key     *rsa.PrivateKey
	subject string
	email   string
	// nonce overrides the nonce claim; empty means echo the login nonce
	nonce string
	// loginNonce is the nonce captured from the most recent login redirect
	loginNonce string
}

func newFakeIdP(t *testing.T) *fakeIdP {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	idp := &fakeIdP{key: key, subject: "sub-1", email: "user@example.com"}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                idp.server.URL,
			"authorization_endpoint":                idp.server.URL + "/authorize",
			"token_endpoint":                        idp.server.URL + "/token",
			"jwks_uri":                              idp.server.URL + "/keys",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"},
		}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code_verifier") == "" {
			http.Error(w, "missing code_verifier", http.StatusBadRequest)
			return
		}

		nonce := idp.nonce
		if nonce == "" {
			nonce = idp.loginNonce
		}

		signer, err := jose.NewSigner(
			jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: "k1"}},
			(&jose.SignerOptions{}).WithType("JWT"),
		)
		require.NoError(t, err)

		raw, err := jwt.Signed(signer).Claims(jwt.Claims{
			Issuer:   idp.server.URL,
			Subject:  idp.subject,
			Audience: jwt.Audience{"client-id"},
			Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		}).Claims(map[string]any{
			"nonce":          nonce,
			"email":          idp.email,
			"email_verified": true,
		}).Serialize()
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access",
			"token_type":   "Bearer",
			"id_token":     raw,
		})
	})

	idp.server = httptest.NewServer(mux)
	t.Cleanup(idp.server.Close)

	return idp
}

func newOIDCService(t *testing.T, idp *fakeIdP, cfg OIDCConfig) (*Service, http.Handler) {
	t.Helper()

	store := memory.New()
	tokenProvider := &TokenProvider{Store: store}
	userProvider := &UserProvider{Store: store}
	require.NoError(t, tokenProvider.Initialize())
	require.NoError(t, userProvider.Initialize())

	cfg.Enabled = true
	cfg.Issuer = idp.server.URL
	cfg.ClientId = "client-id"
	cfg.ClientSecret = "secret"
	cfg.RedirectURL = "http://speakerbob.test/auth/oidc/callback/"

	svc := &Service{TokenProvider: tokenProvider, UserProvider: userProvider, OIDC: cfg}
	router := mux.NewRouter()
	svc.RegisterRoutes(router.PathPrefix("/auth").Subrouter())

	return svc, router
}

func get(h http.Handler, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

// login performs the login redirect and returns the state to use in the callback.
func login(t *testing.T, h http.Handler, idp *fakeIdP) string {
	t.Helper()

	rec := get(h, "/auth/oidc/login/")
	require.Equal(t, http.StatusFound, rec.Code)

	loc, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)

	q := loc.Query()
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.NotEmpty(t, q.Get("code_challenge"))
	assert.Equal(t, "client-id", q.Get("client_id"))
	require.NotEmpty(t, q.Get("state"))
	require.NotEmpty(t, q.Get("nonce"))

	idp.loginNonce = q.Get("nonce")

	return q.Get("state")
}

func TestOIDCLoginRedirect(t *testing.T) {
	idp := newFakeIdP(t)
	_, h := newOIDCService(t, idp, OIDCConfig{})

	rec := get(h, "/auth/oidc/login/")

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), idp.server.URL+"/authorize")
}

func TestOIDCCallbackCreatesUserAndSessionToken(t *testing.T) {
	idp := newFakeIdP(t)
	svc, h := newOIDCService(t, idp, OIDCConfig{})

	state := login(t, h, idp)
	rec := get(h, "/auth/oidc/callback/?code=abc&state="+state)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location"))

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, cookieName, cookies[0].Name)

	user := svc.UserProvider.GetByEmail("user@example.com")
	require.NotNil(t, user)
	assert.Equal(t, []Principal{NewPrincipal("oidc", "sub-1")}, user.Principals)

	token := svc.TokenProvider.GetByToken(cookies[0].Value)
	require.NotNil(t, token)
	assert.Equal(t, Session, token.Type)
	assert.Equal(t, user.Id, token.UserId)
}

func TestOIDCCallbackBindsPrincipalToExistingEmailUser(t *testing.T) {
	idp := newFakeIdP(t)
	svc, h := newOIDCService(t, idp, OIDCConfig{})

	existing := NewUser()
	existing.Email = "user@example.com"
	require.NoError(t, svc.UserProvider.Save(&existing))

	state := login(t, h, idp)
	rec := get(h, "/auth/oidc/callback/?code=abc&state="+state)
	require.Equal(t, http.StatusFound, rec.Code)

	user := svc.UserProvider.GetByEmail("user@example.com")
	require.NotNil(t, user)
	assert.Equal(t, existing.Id, user.Id)
	assert.Equal(t, []Principal{NewPrincipal("oidc", "sub-1")}, user.Principals)
}

func TestOIDCCallbackUnknownStateUnauthorized(t *testing.T) {
	idp := newFakeIdP(t)
	_, h := newOIDCService(t, idp, OIDCConfig{})

	rec := get(h, "/auth/oidc/callback/?code=abc&state=nope")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestOIDCCallbackStateIsSingleUse(t *testing.T) {
	idp := newFakeIdP(t)
	_, h := newOIDCService(t, idp, OIDCConfig{})

	state := login(t, h, idp)
	require.Equal(t, http.StatusFound, get(h, "/auth/oidc/callback/?code=abc&state="+state).Code)

	assert.Equal(t, http.StatusUnauthorized, get(h, "/auth/oidc/callback/?code=abc&state="+state).Code)
}

func TestOIDCCallbackNonceMismatchUnauthorized(t *testing.T) {
	idp := newFakeIdP(t)
	svc, h := newOIDCService(t, idp, OIDCConfig{})
	idp.nonce = "some-other-nonce"

	state := login(t, h, idp)
	rec := get(h, "/auth/oidc/callback/?code=abc&state="+state)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, rec.Result().Cookies())
	assert.Nil(t, svc.UserProvider.GetByEmail("user@example.com"))
}

func TestOIDCCallbackDisallowedEmailDomain(t *testing.T) {
	idp := newFakeIdP(t)
	svc, h := newOIDCService(t, idp, OIDCConfig{AllowedEmailDomains: []string{"corp.test"}})

	state := login(t, h, idp)
	rec := get(h, "/auth/oidc/callback/?code=abc&state="+state)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/permission-denied/", rec.Header().Get("Location"))
	assert.Empty(t, rec.Result().Cookies())
	assert.Nil(t, svc.UserProvider.GetByEmail("user@example.com"))
}
