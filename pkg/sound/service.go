package sound

import (
	"context"
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const soundCreateMessageType websocket.MessageType = "sound.sound.create"
const groupCreateMessageType websocket.MessageType = "sound.group.create"

type Service struct {
	SoundProvider    *SoundProvider
	GroupProvider    *GroupProvider
	WebsocketService *websocket.Service
	MaxSoundDuration time.Duration
}

const cleanupInterval = 4 * time.Hour
const hiddenSoundTTL = 24 * time.Hour

func (s *Service) RegisterRoutes(router *mux.Router) {
	// sounds
	router.HandleFunc("/sound/search/", s.list).Methods("GET")
	router.HandleFunc("/sound/", s.createSound).Methods("POST")
	router.HandleFunc("/sound/{soundId}/", s.updateSound).Methods("PATCH")
	router.HandleFunc("/sound/{soundId}/", s.deleteSound).Methods("DELETE")
	router.HandleFunc("/sound/{soundId}/download/", s.downloadSound).Methods("GET")

	// groups
	router.HandleFunc("/group/", s.createGroup).Methods("POST")
	router.HandleFunc("/group/{groupId}/", s.deleteGroup).Methods("DELETE")
}

func (s *Service) Run(ctx context.Context) {
	var err error
	var now time.Time
	var ticker *time.Ticker

	ticker = time.NewTicker(cleanupInterval)

	logrus.Info("starting sound service worker")
	for {
		select {
		case <-ctx.Done():
			break
		case <-ticker.C:
			logrus.Debug("starting hidden sound cleanup")
			now = time.Now()

			for _, sound := range s.SoundProvider.List() {
				if sound.Hidden && now.Sub(sound.CreatedAt) > hiddenSoundTTL {
					logrus.Infof("deleting \"%s\" expired hidden sounds", sound.Id)
					err = s.SoundProvider.Delete(sound)
					if err != nil {
						logrus.Errorf("error deleting hidden sound: %d", err)
					}
				}
			}
		}
	}
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) {
	var err error
	var sounds []*Sound
	var groups []*Group

	q := r.URL.Query().Get("q")

	sounds = s.SoundProvider.Search(q)
	groups = s.GroupProvider.Search(q)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{"sounds": sounds, "groups": groups})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) createSound(w http.ResponseWriter, r *http.Request) {
	var sound *Sound

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	for _, fileHeaderArray := range r.MultipartForm.File {
		fileHeader := fileHeaderArray[0]
		data, err := fileHeader.Open()
		if err != nil {
			return // upload aborted
		}

		sound, err = s.SoundProvider.NewSound(fileHeader.Filename, data, s.MaxSoundDuration)
		if err != nil {
			logrus.Errorf("failed to create sound: %s  -- %s", fileHeader.Filename, err.Error())
			w.WriteHeader(http.StatusInternalServerError) // TODO: determine file is bad or upload is bad
			return
		}
		break
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(sound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) updateSound(w http.ResponseWriter, r *http.Request) {
	var currentSound *Sound
	var sound Sound
	var err error

	soundId := mux.Vars(r)["soundId"]

	// decode user request
	err = json.NewDecoder(r.Body).Decode(&sound)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// names cannot be set to empty
	if sound.Name == "" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// load existing values
	currentSound = s.SoundProvider.Get(soundId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if currentSound.Name != "" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// write user changes
	currentSound.Name = sound.Name
	currentSound.Hidden = false
	err = s.SoundProvider.Save(currentSound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.WebsocketService.BroadcastMessage(soundCreateMessageType, currentSound)
}

func (s *Service) deleteSound(w http.ResponseWriter, r *http.Request) {
	var sound Sound

	sound.Id = mux.Vars(r)["soundId"]

	err := DeleteSoundWithGroups(s.GroupProvider, s.SoundProvider, &sound)
	if err != nil && err != mux.ErrNotFound {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) downloadSound(w http.ResponseWriter, r *http.Request) {
	var sound Sound
	var err error

	sound.Id = mux.Vars(r)["soundId"]

	w.Header().Set("Content-Type", "audio/mp3")
	w.Header().Set("Cache-Control", "max-age=2592000s")

	err = s.SoundProvider.ReadAudio(&sound, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err == badger.ErrKeyNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) createGroup(w http.ResponseWriter, r *http.Request) {
	var err error

	var group Group
	var userGroup Group

	// decode user request
	err = json.NewDecoder(r.Body).Decode(&userGroup)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// names cannot be set to empty
	if userGroup.Name == "" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// validate we have at least 2 sounds
	if len(userGroup.SoundIds) < 2 {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// validate sound ids
	for i := range group.SoundIds {
		if s.SoundProvider.Get(group.SoundIds[i]).Id != "" {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
	}

	group = NewGroup()
	group.Name = userGroup.Name
	group.SoundIds = userGroup.SoundIds

	// create group
	err = s.GroupProvider.Save(&group)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.WebsocketService.BroadcastMessage(groupCreateMessageType, &group)
}

func (s *Service) deleteGroup(w http.ResponseWriter, r *http.Request) {
	var group Group

	group.Id = mux.Vars(r)["groupId"]

	err := s.GroupProvider.Delete(&group)

	if err != nil && err != mux.ErrNotFound {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}
