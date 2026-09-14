package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	"github.com/paynejacob/speakerbob/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubAuthProvider struct{}

func (stubAuthProvider) Name() string { return "stub" }
func (stubAuthProvider) VerifyCallback(*http.Request) (auth.Principal, string, error) {
	return "", "", nil
}
func (stubAuthProvider) LoginRedirect(http.ResponseWriter, *http.Request, string) {}

func newTestAuthService(t *testing.T) *auth.Service {
	t.Helper()

	tokenProvider := &auth.TokenProvider{Store: memory.New()}
	require.NoError(t, tokenProvider.Initialize())

	userProvider := &auth.UserProvider{Store: memory.New()}
	require.NoError(t, userProvider.Initialize())

	return &auth.Service{
		TokenProvider: tokenProvider,
		UserProvider:  userProvider,
		Providers:     []auth.Provider{stubAuthProvider{}},
	}
}

func newTestUserAndToken(t *testing.T, authService *auth.Service, email string) *auth.Token {
	t.Helper()

	user := auth.NewUser()
	user.Email = email
	require.NoError(t, authService.UserProvider.Save(&user))

	token := auth.NewToken()
	token.Type = auth.Websocket
	token.UserId = user.Id
	require.NoError(t, authService.TokenProvider.Save(&token))

	return &token
}

// fakeConn observes broadcasts sent to it without a real network connection.
func newFakeConn(service *Service, token *auth.Token) *Conn {
	return &Conn{service: service, token: token, send: make(chan interface{}, sendChannelSize)}
}

// registerAndReadPresence registers conn on svc and returns the PresenceMessage
// broadcast that follows — every register/unregister sends exactly two
// messages, in order: ConnectionCountMessage then PresenceMessage.
func registerAndReadPresence(t *testing.T, svc *Service, conn *Conn, observer *Conn) PresenceMessage {
	t.Helper()
	svc.registerConnection(conn)
	return readPresence(t, observer)
}

func unregisterAndReadPresence(t *testing.T, svc *Service, conn *Conn, observer *Conn) PresenceMessage {
	t.Helper()
	svc.unRegisterConnection(conn)
	return readPresence(t, observer)
}

func readPresence(t *testing.T, observer *Conn) PresenceMessage {
	t.Helper()
	_, ok := (<-observer.send).(ConnectionCountMessage)
	require.True(t, ok, "expected ConnectionCountMessage first")
	msg, ok := (<-observer.send).(PresenceMessage)
	require.True(t, ok, "expected PresenceMessage second")
	return msg
}

func TestBroadcastPresenceOnConnect(t *testing.T) {
	authService := newTestAuthService(t)
	token := newTestUserAndToken(t, authService, "alice@example.com")

	svc := &Service{AuthService: authService}

	observer := newFakeConn(svc, nil)
	registerAndReadPresence(t, svc, observer, observer)

	msg := registerAndReadPresence(t, svc, newFakeConn(svc, token), observer)
	assert.Equal(t, MessageType(PresenceMessageType), msg.Type)
	assert.Equal(t, []string{"alice"}, msg.Users)
}

func TestBroadcastPresenceDedupesMultipleConnectionsFromSameUser(t *testing.T) {
	authService := newTestAuthService(t)
	token := newTestUserAndToken(t, authService, "bob@example.com")

	svc := &Service{AuthService: authService}

	observer := newFakeConn(svc, nil)
	registerAndReadPresence(t, svc, observer, observer)

	firstPresence := registerAndReadPresence(t, svc, newFakeConn(svc, token), observer)
	assert.Equal(t, []string{"bob"}, firstPresence.Users)

	secondPresence := registerAndReadPresence(t, svc, newFakeConn(svc, token), observer)
	assert.Equal(t, []string{"bob"}, secondPresence.Users, "same user connected twice should appear once")
}

func TestBroadcastPresenceOnDisconnect(t *testing.T) {
	authService := newTestAuthService(t)
	token := newTestUserAndToken(t, authService, "carol@example.com")

	svc := &Service{AuthService: authService}

	observer := newFakeConn(svc, nil)
	registerAndReadPresence(t, svc, observer, observer)

	conn := newFakeConn(svc, token)
	registerAndReadPresence(t, svc, conn, observer)

	msg := unregisterAndReadPresence(t, svc, conn, observer)
	assert.Empty(t, msg.Users)
}

func TestBroadcastPresenceSkipsUnauthenticatedConnections(t *testing.T) {
	authService := newTestAuthService(t)

	svc := &Service{AuthService: authService}

	observer := newFakeConn(svc, nil)
	registerAndReadPresence(t, svc, observer, observer)

	// nil token: auth disabled, or otherwise unresolved identity.
	msg := registerAndReadPresence(t, svc, newFakeConn(svc, nil), observer)
	assert.Empty(t, msg.Users)
}

// sanity check that connect() itself still threads the resolved token through
// to the registered Conn, exercised via a real HTTP + websocket handshake.
func TestConnectThreadsTokenIntoConn(t *testing.T) {
	authService := newTestAuthService(t)
	token := newTestUserAndToken(t, authService, "dave@example.com")

	svc := &Service{AuthService: authService}
	router := httptest.NewServer(http.HandlerFunc(svc.connect))
	defer router.Close()

	wsURL := "ws" + router.URL[len("http"):] + "?token=" + token.Token
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = ws.Close() }()

	require.Eventually(t, func() bool {
		svc.m.RLock()
		defer svc.m.RUnlock()
		return len(svc.connections) == 1 && svc.connections[0].token != nil && svc.connections[0].token.UserId == token.UserId
	}, time.Second, 10*time.Millisecond)
}
