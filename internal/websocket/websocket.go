package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"net/http"
	"regexp"
	"speakerbob/internal/authentication"
)

var backendProtoRegex = regexp.MustCompile("(.*)://.*?")

func NewService(backendURL string, db *gorm.DB) *Service {
	var backend Backend
	var upgrader = websocket.Upgrader{}

	if matches := backendProtoRegex.FindStringSubmatch(backendURL); len(matches) > 0 {
		proto := matches[0]
		switch proto {
		case "memory":
			backend = NewMemoryBackend()
		default:
			panic(fmt.Sprintf("\"%s\" is not a valid authentication backend proto", proto))
		}
	}

	return &Service{backend, upgrader, db}
}

type Service struct {
	backend  Backend
	upgrader websocket.Upgrader

	db *gorm.DB
}

func (s *Service) WSConnect(w http.ResponseWriter, r *http.Request) {
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

	var user authentication.User
	if err := s.db.Where("id = ?", r.Context().Value("userId")).First(&user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	connection := s.backend.RegisterConnection(ws, user, channels, nsfw)

	defer ws.Close()
	s.backend.CloseConnection(connection)
}

func (s *Service) WSMessageConsumer() {
	for {
		for message := range s.backend.Channel() {
			for _, connection := range s.backend.LocalConnections(message.Channels) {
				if err := connection.ws.WriteJSON(message); err != nil {
					s.backend.CloseConnection(connection)
				}
			}
		}
	}
}
