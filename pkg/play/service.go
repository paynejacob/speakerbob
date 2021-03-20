package play

import (
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Service struct {
	playQueue *queue

	soundStore       *sound.Provider
	websocketService *websocket.Service
}

func NewService(soundStore *sound.Provider, websocketService *websocket.Service) *Service {
	return &Service{playQueue: newQueue(), websocketService: websocketService, soundStore: soundStore}
}

func (s *Service) RegisterRoutes(parent *mux.Router, prefix string) {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/sound/{soundId}/", s.playSound).Methods("PUT")
}

func (s *Service) Run() {
	logrus.Info("starting play service worker")

	s.playQueue.ConsumeQueue(s.websocketService)
}

func (s *Service) playSound(w http.ResponseWriter, r *http.Request) {
	var _sound sound.Sound
	var err error

	_sound.Id = mux.Vars(r)["soundId"]

	err = s.soundStore.GetSound(&_sound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.playQueue.EnqueueSound(_sound)

	w.WriteHeader(http.StatusAccepted)
}
