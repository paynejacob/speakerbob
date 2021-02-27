package websocket

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var upgrader websocket.Upgrader

const ConnectionCountMessageType = "connection_count"

type ConnectionCountMessagePayload struct {
	Count int `json:"count"`
}

type Service struct {
	m sync.RWMutex

	connections []*Conn
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) RegisterRoutes(parent *mux.Router, prefix string) {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/", s.connect).Methods("GET")
}

func (s *Service) BroadcastMessage(t MessageType, payload interface{}) {
	s.m.RLock()
	defer s.m.RUnlock()

	message := Message{
		Type:    t,
		Payload: payload,
	}

	for i := range s.connections {
		s.connections[i].SendMessage(message)
	}
}

func (s *Service) Run() {}

func (s *Service) connect(w http.ResponseWriter, r *http.Request) {
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

	s.BroadcastMessage(ConnectionCountMessageType, ConnectionCountMessagePayload{Count: connectionCount})
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

	s.BroadcastMessage(ConnectionCountMessageType, ConnectionCountMessagePayload{Count: connectionCount})
}
