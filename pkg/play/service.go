package play

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Service struct {
	playQueue *queue

	soundProvider    *sound.Provider
	websocketService *websocket.Service
}

func NewService(soundStore *sound.Provider, websocketService *websocket.Service) *Service {
	return &Service{playQueue: newQueue(), websocketService: websocketService, soundProvider: soundStore}
}

func (s *Service) RegisterRoutes(parent *mux.Router, prefix string) {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/sound/{soundId}/", s.playSound).Methods("PUT")
	router.HandleFunc("/group/{groupId}/", s.playGroup).Methods("PUT")
	router.HandleFunc("/say/", s.say).Methods("PUT")
}

func (s *Service) Run() {
	logrus.Info("starting play service worker")

	s.playQueue.ConsumeQueue(s.websocketService)
}

func (s *Service) playSound(w http.ResponseWriter, r *http.Request) {
	var _sound sound.Sound
	var err error

	_sound.Id = mux.Vars(r)["soundId"]

	err = s.soundProvider.GetSound(&_sound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.playQueue.EnqueueSounds(_sound)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) playGroup(w http.ResponseWriter, r *http.Request) {
	var group sound.Group
	var err error

	group.Id = mux.Vars(r)["groupId"]

	err = s.soundProvider.GetGroup(&group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sounds := make([]sound.Sound, len(group.SoundIds))
	for i := range group.SoundIds {
		sounds[i].Id = group.SoundIds[i]
		err = s.soundProvider.GetSound(&sounds[i])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	s.playQueue.EnqueueSounds(sounds...)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) say(w http.ResponseWriter, r *http.Request) {
	var err error
	var text string
	var _sound sound.Sound

	// parse user request
	err = json.NewDecoder(r.Body).Decode(&text)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// generate sound
	_sound, err = s.soundProvider.CreateTextToSpeechSound(text)
	if err != nil {
		logrus.Info(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// enqueue playback
	s.playQueue.EnqueueSounds(_sound)

	w.WriteHeader(http.StatusAccepted)
}
