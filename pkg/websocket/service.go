package websocket

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/paynejacob/speakerbob/pkg/auth"
	"net/http"
	"sort"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Service struct {
	AuthService *auth.Service

	// Broadcaster fans BroadcastMessage out to connected clients. Nil defaults
	// to in-process fan-out, so existing zero-value Service{...} construction
	// keeps working unchanged.
	Broadcaster Broadcaster

	m           sync.RWMutex
	connections []*Conn
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ws/", s.connect).Methods("GET")
}

func (s *Service) broadcaster() Broadcaster {
	if s.Broadcaster != nil {
		return s.Broadcaster
	}
	return &inProcessBroadcaster{service: s}
}

func (s *Service) BroadcastMessage(msg interface{}) {
	s.broadcaster().Broadcast(msg)
}

func (s *Service) Run(context.Context) {}

func (s *Service) connect(w http.ResponseWriter, r *http.Request) {
	token, valid := s.AuthService.VerifyWebsocket(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	conn := NewConn(ws, s, token)
	s.registerConnection(conn)

	go conn.writePump()
	conn.readPump()
}

func (s *Service) registerConnection(conn *Conn) {
	var connectionCount int

	s.m.Lock()
	s.connections = append(s.connections, conn)
	connectionCount = len(s.connections)
	s.m.Unlock()

	s.BroadcastMessage(ConnectionCountMessage{
		Type:  ConnectionCountMessageType,
		Count: connectionCount,
	})

	s.broadcastPresence()
}

func (s *Service) unRegisterConnection(conn *Conn) {
	var connectionCount int

	s.m.Lock()
	for i := 0; i < len(s.connections); i++ {
		if s.connections[i] == conn {
			s.connections = append(s.connections[:i], s.connections[i+1:]...)
			break
		}
	}
	connectionCount = len(s.connections)
	s.m.Unlock()

	s.BroadcastMessage(ConnectionCountMessage{
		Type:  ConnectionCountMessageType,
		Count: connectionCount,
	})

	s.broadcastPresence()
}

// broadcastPresence sends the deduplicated list of signed-in users currently
// connected. Identity is the local-part of the user's email (e.g. "alice"
// from "alice@example.com") rather than the full address: User.Email is
// otherwise treated as non-public (json:"-") throughout this codebase, and
// this broadcast reaches every connected client, not just the user it
// describes.
func (s *Service) broadcastPresence() {
	s.m.RLock()
	seen := make(map[string]struct{}, len(s.connections))
	identities := make([]string, 0, len(s.connections))
	for _, c := range s.connections {
		identity := s.userIdentity(c.token)
		if identity == "" {
			continue
		}
		if _, ok := seen[identity]; ok {
			continue
		}
		seen[identity] = struct{}{}
		identities = append(identities, identity)
	}
	s.m.RUnlock()

	sort.Strings(identities)

	s.BroadcastMessage(PresenceMessage{
		Type:  PresenceMessageType,
		Users: identities,
	})
}

func (s *Service) userIdentity(token *auth.Token) string {
	if token == nil || s.AuthService == nil || s.AuthService.UserProvider == nil {
		return ""
	}

	user := s.AuthService.UserProvider.Get(token.UserId)
	if user == nil || user.Email == "" {
		return ""
	}

	if i := strings.Index(user.Email, "@"); i > 0 {
		return user.Email[:i]
	}

	return user.Email
}
