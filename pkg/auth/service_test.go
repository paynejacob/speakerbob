package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	"github.com/stretchr/testify/assert"
)

// stubProvider satisfies the Provider interface so Service.Enabled() is true
// without needing a real OAuth flow, which this test doesn't exercise.
type stubProvider struct{}

func (stubProvider) Name() string { return "stub" }

func (stubProvider) VerifyCallback(*http.Request) (Principal, string, error) {
	return "", "", nil
}

func (stubProvider) LoginRedirect(http.ResponseWriter, *http.Request, string) {}

func TestTokenCleanup(t *testing.T) {
	tokenProvider := &TokenProvider{Store: memory.New()}
	_ = tokenProvider.Initialize()

	svc := &Service{
		TokenProvider: tokenProvider,
		Providers:     []Provider{stubProvider{}},
	}

	svc.RegisterRoutes(mux.NewRouter())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.Run(ctx)

	token := NewToken()
	token.Type = Session
	token.ExpiresAt = time.Now().Add(50 * time.Millisecond)
	_ = tokenProvider.Save(&token)

	svc.enqueueCleanup(cleanupItem{id: token.Id, expiresAt: token.ExpiresAt})

	assert.NotNil(t, tokenProvider.Get(token.Id), "token should exist immediately after creation")

	assert.Eventually(t, func() bool {
		return tokenProvider.Get(token.Id) == nil
	}, 2*time.Second, 10*time.Millisecond, "expected expired token to be cleaned up shortly after its TTL")
}

func TestTokenCleanupSkipsNonExpiringTokens(t *testing.T) {
	tokenProvider := &TokenProvider{Store: memory.New()}
	_ = tokenProvider.Initialize()

	svc := &Service{
		TokenProvider: tokenProvider,
		Providers:     []Provider{stubProvider{}},
	}

	svc.RegisterRoutes(mux.NewRouter())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.Run(ctx)

	// Bearer tokens never set ExpiresAt (zero value) and must never be
	// scheduled for cleanup.
	token := NewToken()
	token.Type = Bearer
	_ = tokenProvider.Save(&token)

	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, tokenProvider.Get(token.Id), "non-expiring token must not be cleaned up")
}
