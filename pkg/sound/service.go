package sound

import (
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
	soundProvider    *SoundProvider
	groupProvider    *GroupProvider
	websocketService *websocket.Service
	maxSoundDuration time.Duration
}

const cleanupInterval = 4 * time.Hour
const hiddenSoundTTL = 24 * time.Hour

func NewService(SoundProvider *SoundProvider, groupProvider *GroupProvider, websocketService *websocket.Service, maxSoundDuration time.Duration) *Service {
	return &Service{
		soundProvider:    SoundProvider,
		groupProvider:    groupProvider,
		websocketService: websocketService,
		maxSoundDuration: maxSoundDuration,
	}
}

func (s *Service) RegisterRoutes(parent *mux.Router, prefix string) {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/", s.list).Methods("GET")

	// sounds
	router.HandleFunc("/sound/", s.createSound).Methods("POST")
	router.HandleFunc("/sound/{soundId}/", s.updateSound).Methods("PATCH")
	router.HandleFunc("/sound/{soundId}/", s.deleteSound).Methods("DELETE")
	router.HandleFunc("/sound/{soundId}/download/", s.downloadSound).Methods("GET")

	// groups
	router.HandleFunc("/group/", s.createGroup).Methods("POST")
	router.HandleFunc("/group/{groupId}/", s.deleteGroup).Methods("DELETE")
}

func (s *Service) Run() {
	var err error
	var now time.Time

	logrus.Info("starting sound service worker")
	for range time.Tick(cleanupInterval) {
		logrus.Debug("starting hidden sound cleanup")
		now = time.Now()

		for _, sound := range s.soundProvider.List() {
			if sound.Hidden && now.Sub(sound.CreatedAt) > hiddenSoundTTL {
				logrus.Infof("deleting \"%s\" expired hidden sounds", sound.Id)

				err = DeleteSound(s.groupProvider, s.soundProvider, sound)
				if err != nil {
					logrus.Errorf("error deleting hidden sound: %d", err)
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

	sounds = s.soundProvider.Search(q)
	groups = s.groupProvider.Search(q)

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
	var sound Sound

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

		sound, err = s.soundProvider.NewSound(fileHeader.Filename, data, s.maxSoundDuration)
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
	currentSound = s.soundProvider.Get(soundId)
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
	err = s.soundProvider.Save(currentSound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.websocketService.BroadcastMessage(soundCreateMessageType, currentSound)
}

func (s *Service) deleteSound(w http.ResponseWriter, r *http.Request) {
	var sound *Sound

	sound.Id = mux.Vars(r)["soundId"]

	err := DeleteSound(s.groupProvider, s.soundProvider, sound)

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

	err = s.soundProvider.GetAudio(&sound, w)

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
		if s.soundProvider.Get(group.SoundIds[i]).Id != "" {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
	}

	group = NewGroup()
	group.Name = userGroup.Name
	group.SoundIds = userGroup.SoundIds

	// create group
	err = s.groupProvider.Save(&group)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.websocketService.BroadcastMessage(groupCreateMessageType, &group)
}

func (s *Service) deleteGroup(w http.ResponseWriter, r *http.Request) {
	var group Group

	group.Id = mux.Vars(r)["groupId"]

	err := s.groupProvider.Delete(&group)

	if err != nil && err != mux.ErrNotFound {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}
