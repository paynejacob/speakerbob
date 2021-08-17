package play

import (
	"context"
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

	SoundProvider    *sound.SoundProvider
	GroupProvider    *sound.GroupProvider
	WebsocketService *websocket.Service
	MaxSoundDuration time.Duration
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/play/sound/{soundId}/", s.playSound).Methods("PUT")
	router.HandleFunc("/play/group/{groupId}/", s.playGroup).Methods("PUT")
	router.HandleFunc("/play/say/", s.say).Methods("PUT")
}

func (s *Service) Run(ctx context.Context) {
	logrus.Info("starting play service worker")

	s.playQueue.ConsumeQueue(ctx, s.WebsocketService)
}

func (s *Service) playSound(w http.ResponseWriter, r *http.Request) {
	_sound := *s.SoundProvider.Get(mux.Vars(r)["soundId"])
	if _sound.Id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.playQueue.EnqueueSounds(_sound)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) playGroup(w http.ResponseWriter, r *http.Request) {
	group := s.GroupProvider.Get(mux.Vars(r)["groupId"])
	if group.Id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sounds := make([]sound.Sound, len(group.SoundIds))
	for i := range group.SoundIds {
		sounds[i] = *s.SoundProvider.Get(group.SoundIds[i])
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
	_sound, err := s.SoundProvider.NewTTSSound(text, s.MaxSoundDuration)
	if err != nil {
		logrus.Errorf("failed to codegen tts sound: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// enqueue playback
	s.playQueue.EnqueueSounds(*_sound)

	w.WriteHeader(http.StatusAccepted)
}
