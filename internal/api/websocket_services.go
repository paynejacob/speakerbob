package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"net/http"
	"net/url"
)

type WebsocketService struct {
	backend  WebsocketBackend
	upgrader websocket.Upgrader

	db *gorm.DB
}

func NewWebsocketService(backendURL string, db *gorm.DB) *WebsocketService {
	var backend WebsocketBackend
	var upgrader = websocket.Upgrader{}

	parsedUrl, err := url.Parse(backendURL)
	if err != nil {
		panic("invalid backend url")
	}

	switch parsedUrl.Scheme {
	case "memory":
		backend = NewWebsocketMemoryBackend()
	default:
		panic(fmt.Sprintf("\"%s\" is not a valid authentication backend proto", parsedUrl.Scheme))
	}

	return &WebsocketService{backend, upgrader, db}
}

func (s *WebsocketService) RegisterRoutes(parent *mux.Router, prefix string) *mux.Router {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/", s.WSConnect).Methods("GET")

	return router
}

func (s *WebsocketService) WSConnect(w http.ResponseWriter, r *http.Request) {
	var channels ChannelSet

	if rawChannels, ok := r.URL.Query()["channels"]; ok {
		for _, channel := range rawChannels {
			channels.Add(&Channel{channel})
		}
	}

	var nsfw bool
	rawNSFW, ok := r.URL.Query()["nsfw"]
	if !ok {
		nsfw = true
	} else {
		nsfw = rawNSFW[0] == "1"
	}

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var user User
	if err := s.db.Where("id = ?", r.Context().Value("userId")).First(&user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	connection := s.backend.RegisterConnection(ws, user, channels, nsfw)

	defer ws.Close()
	s.backend.CloseConnection(connection)
}

func (s *WebsocketService) SendMessage(message IMessage) {
	s.backend.SendMessage(message)
}

func (s *WebsocketService) WSMessageConsumer() {
	for {
		for message := range s.backend.Channel() {
			for _, connection := range s.backend.LocalConnections(message.Channels()) {
				if err := connection.ws.WriteJSON(message); err != nil {
					s.backend.CloseConnection(connection)
				}
			}
		}
	}
}
