package sound

import (
	"context"
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/service"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type Service struct {
	SoundProvider    *SoundProvider
	GroupProvider    *GroupProvider
	WebsocketService *websocket.Service
	MaxSoundDuration time.Duration

	playQueue playQueue
}

const cleanupInterval = 4 * time.Hour
const hiddenSoundTTL = 24 * time.Hour

func (s *Service) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/sound").Subrouter()

	sounds := r.PathPrefix("/sounds").Subrouter()
	sounds.HandleFunc("/", s.listSound).Methods(http.MethodGet)
	sounds.HandleFunc("/", s.createSound).Methods(http.MethodPost)
	sounds.HandleFunc("/{soundId}/", s.updateSound).Methods(http.MethodPatch)
	sounds.HandleFunc("/{soundId}/", s.deleteSound).Methods(http.MethodDelete)
	sounds.HandleFunc("/{soundId}/play/", s.playSound).Methods(http.MethodPut)
	router.HandleFunc("/{soundId}/download/", s.downloadSound).Methods(http.MethodGet)

	groups := r.PathPrefix("/groups").Subrouter()
	groups.HandleFunc("/", s.listGroup).Methods(http.MethodGet)
	groups.HandleFunc("/", s.createGroup).Methods(http.MethodPost)
	groups.HandleFunc("/{groupId}/", s.updateGroup).Methods(http.MethodPatch)
	groups.HandleFunc("/{groupId}/", s.deleteGroup).Methods(http.MethodDelete)
	groups.HandleFunc("/{groupId}/play/", s.playGroup).Methods(http.MethodPut)

	r.HandleFunc("/search/", s.search).Methods(http.MethodGet)
	r.HandleFunc("/say/", s.say).Methods(http.MethodPut)

}

func (s *Service) Run(ctx context.Context) {
	var err error
	var now time.Time
	var ticker *time.Ticker

	s.playQueue = playQueue{
		m:           sync.RWMutex{},
		playChannel: make(chan bool, 0),
		sounds:      make([]Sound, 0),
	}

	go s.playQueue.ConsumeQueue(ctx, s.WebsocketService)

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

func (s *Service) listSound(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	sounds := make([]*Sound, 0)
	for _, sound := range s.SoundProvider.List() {
		if sound.Hidden {
			continue
		}

		sounds = append(sounds, sound)
	}

	_ = json.NewEncoder(w).Encode(sounds)
}

func (s *Service) createSound(w http.ResponseWriter, r *http.Request) {
	var sound *Sound

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("invalid request format"))
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
			service.WriteErrorResponse(w, err)
			return
		}
		break
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sound)
}

func (s *Service) updateSound(w http.ResponseWriter, r *http.Request) {
	var sound *Sound
	var requestSound Sound
	var err error

	// load existing values
	sound = s.SoundProvider.Get(mux.Vars(r)["soundId"])
	if sound == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// decode user request
	err = json.NewDecoder(r.Body).Decode(&requestSound)
	if err != nil {
		service.WriteErrorResponse(w, service.NewNotAcceptableError(err.Error()))
		return
	}

	// names cannot be set to empty
	if !(0 < len(requestSound.Name) && len(requestSound.Name) < 30) {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("requestSound names must be between 1 and 15 characters"))
		return
	}

	// write user changes
	sound.Name = requestSound.Name
	sound.Hidden = false
	err = s.SoundProvider.Save(sound)
	if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}

	s.WebsocketService.BroadcastMessage(SoundMessage{
		Type:  websocket.UpdateSoundMessageType,
		Sound: sound,
	})

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) deleteSound(w http.ResponseWriter, r *http.Request) {
	var sound *Sound

	sound = s.SoundProvider.Get(mux.Vars(r)["soundId"])
	if sound == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := DeleteSoundWithGroups(s.GroupProvider, s.SoundProvider, sound)
	if err != nil && err != mux.ErrNotFound {
		service.WriteErrorResponse(w, err)
		return
	}

	s.WebsocketService.BroadcastMessage(SoundMessage{
		Type:  websocket.DeleteSoundMessageType,
		Sound: sound,
	})

	w.WriteHeader(http.StatusNoContent)
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

func (s *Service) downloadSound(w http.ResponseWriter, r *http.Request) {
	var sound Sound
	var err error

	sound.Id = mux.Vars(r)["soundId"]

	w.Header().Set("Content-Type", "audio/mp3")
	w.Header().Set("Cache-Control", "max-age=1y")

	err = s.SoundProvider.ReadAudio(&sound, w)
	if err != nil {
		service.WriteErrorResponse(w, err)
	}

	if err == badger.ErrKeyNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}
}

func (s *Service) listGroup(w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(s.GroupProvider.List())
}

func (s *Service) createGroup(w http.ResponseWriter, r *http.Request) {
	var err error

	var group Group
	var requestGroup Group

	// decode user request
	err = json.NewDecoder(r.Body).Decode(&requestGroup)
	if err != nil {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("unable to parse request"))
		return
	}

	// names cannot be set to empty
	if !(0 < len(requestGroup.Name) && len(requestGroup.Name) < 30) {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("sound names must be between 1 and 15 characters"))
		return
	}

	// validate we have at least 2 sounds
	if len(requestGroup.SoundIds) < 2 {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("groups must consist of 2 or more sounds"))
		return
	}

	// validate sound ids
	for i := range requestGroup.SoundIds {
		if s.SoundProvider.Get(group.SoundIds[i]).Id != "" {
			service.WriteErrorResponse(w, service.NewNotAcceptableError("invalid sound id: "+requestGroup.SoundIds[i]))
			return
		}
	}

	group = NewGroup()
	group.Name = requestGroup.Name
	group.SoundIds = requestGroup.SoundIds

	// create group
	err = s.GroupProvider.Save(&group)
	if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&group)
	if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	s.WebsocketService.BroadcastMessage(GroupMessage{
		Type:  websocket.CreateGroupMessageType,
		Group: &group,
	})
}

func (s *Service) updateGroup(w http.ResponseWriter, r *http.Request) {
	var err error

	var group *Group
	var requestGroup Group

	group = s.GroupProvider.Get(mux.Vars(r)["groupId"])
	if group.Id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// decode user request
	err = json.NewDecoder(r.Body).Decode(&requestGroup)
	if err != nil {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("unable to parse request"))
		return
	}

	// names cannot be set to empty
	if !(0 < len(requestGroup.Name) && len(requestGroup.Name) < 30) {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("sound names must be between 1 and 15 characters"))
		return
	}

	// validate we have at least 2 sounds
	if len(requestGroup.SoundIds) < 2 {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("groups must consist of 2 or more sounds"))
		return
	}

	// validate sound ids
	for i := range requestGroup.SoundIds {
		if s.SoundProvider.Get(group.SoundIds[i]).Id != "" {
			service.WriteErrorResponse(w, service.NewNotAcceptableError("invalid sound id: "+requestGroup.SoundIds[i]))
			return
		}
	}

	group.Name = requestGroup.Name
	group.SoundIds = requestGroup.SoundIds

	// create group
	err = s.GroupProvider.Save(group)
	if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	s.WebsocketService.BroadcastMessage(GroupMessage{
		Type:  websocket.CreateGroupMessageType,
		Group: group,
	})
}

func (s *Service) deleteGroup(w http.ResponseWriter, r *http.Request) {
	var group Group

	group.Id = mux.Vars(r)["groupId"]

	err := s.GroupProvider.Delete(&group)

	if err != nil && err != mux.ErrNotFound {
		service.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) playGroup(w http.ResponseWriter, r *http.Request) {
	group := s.GroupProvider.Get(mux.Vars(r)["groupId"])
	if group.Id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sounds := make([]Sound, len(group.SoundIds))
	for i := range group.SoundIds {
		sounds[i] = *s.SoundProvider.Get(group.SoundIds[i])
	}

	s.playQueue.EnqueueSounds(sounds...)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) search(w http.ResponseWriter, r *http.Request) {
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

func (s *Service) say(w http.ResponseWriter, r *http.Request) {
	var text string

	// parse user request
	err := json.NewDecoder(r.Body).Decode(&text)
	if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}

	// generate sound
	sound, err := s.SoundProvider.NewTTSSound(text, s.MaxSoundDuration)
	if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}

	// enqueue playback
	s.playQueue.EnqueueSounds(*sound)

	w.WriteHeader(http.StatusAccepted)
}
