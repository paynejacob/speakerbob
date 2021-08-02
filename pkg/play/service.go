package play

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Service struct {
	playQueue *queue

	soundProvider    *sound.SoundProvider
	groupProvider    *sound.GroupProvider
	websocketService *websocket.Service
	maxSoundDuration time.Duration
}

func NewService(soundProvider *sound.SoundProvider, groupProvider *sound.GroupProvider, websocketService *websocket.Service, maxSoundDuration time.Duration) *Service {
	return &Service{playQueue: newQueue(), websocketService: websocketService, soundProvider: soundProvider, groupProvider: groupProvider, maxSoundDuration: maxSoundDuration}
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
	_sound := *s.soundProvider.Get(mux.Vars(r)["soundId"])
	if _sound.Id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.playQueue.EnqueueSounds(_sound)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) playGroup(w http.ResponseWriter, r *http.Request) {
	group := s.groupProvider.Get(mux.Vars(r)["groupId"])
	if group.Id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sounds := make([]sound.Sound, len(group.SoundIds))
	for i := range group.SoundIds {
		sounds[i] = *s.soundProvider.Get(group.SoundIds[i])
	}

	s.playQueue.EnqueueSounds(sounds...)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) say(w http.ResponseWriter, r *http.Request) {
	var text string

	// parse user request
	if json.NewDecoder(r.Body).Decode(&text) != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// codegen sound
	_sound, err := s.soundProvider.NewTTSSound(text, s.maxSoundDuration)
	if err != nil {
		logrus.Errorf("failed to codegen tts sound: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// enqueue playback
	s.playQueue.EnqueueSounds(*_sound)

	w.WriteHeader(http.StatusAccepted)
}
