package websocket

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/paynejacob/speakerbob/pkg/auth"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Service struct {
	AuthService *auth.Service

	m           sync.RWMutex
	connections []*Conn
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ws/", s.connect).Methods("GET")
}

func (s *Service) BroadcastMessage(msg interface{}) {
	s.m.RLock()

	for i := range s.connections {
		s.connections[i].SendMessage(msg)
	}

	s.m.RUnlock()
}

func (s *Service) Run(context.Context) {}

func (s *Service) connect(w http.ResponseWriter, r *http.Request) {
	if _, valid := s.AuthService.VerifyWebsocket(r); !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	conn := NewConn(ws, s)
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
}
