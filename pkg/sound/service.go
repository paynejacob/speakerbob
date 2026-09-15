package sound

import (
	"container/heap"
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/service"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
)

type Service struct {
	SoundProvider    *SoundProvider
	GroupProvider    *GroupProvider
	WebsocketService *websocket.Service
	MaxSoundDuration time.Duration
	HiddenSoundTTL   time.Duration

	playQueue playQueue
	cleanupCh chan cleanupItem
}

const hiddenSoundTTL = 24 * time.Hour
const cleanupChannelBufferSize = 64

func (s *Service) RegisterRoutes(router *mux.Router) {
	if s.HiddenSoundTTL <= 0 {
		s.HiddenSoundTTL = hiddenSoundTTL
	}
	s.cleanupCh = make(chan cleanupItem, cleanupChannelBufferSize)

	r := router.PathPrefix("/sound").Subrouter()

	sounds := r.PathPrefix("/sounds").Subrouter()
	sounds.HandleFunc("/", s.listSound).Methods(http.MethodGet)
	sounds.HandleFunc("/", s.createSound).Methods(http.MethodPost)
	sounds.HandleFunc("/{soundId}/", s.getSound).Methods(http.MethodGet)
	sounds.HandleFunc("/{soundId}/", s.updateSound).Methods(http.MethodPatch)
	sounds.HandleFunc("/{soundId}/", s.deleteSound).Methods(http.MethodDelete)
	sounds.HandleFunc("/{soundId}/play/", s.playSound).Methods(http.MethodPut)
	sounds.HandleFunc("/{soundId}/download/", s.downloadSound).Methods(http.MethodGet)

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
	s.playQueue = playQueue{
		m:           sync.RWMutex{},
		playChannel: make(chan bool, 0),
		sounds:      make([]Sound, 0),
	}

	go s.playQueue.ConsumeQueue(ctx, s.WebsocketService, s.SoundProvider)

	pending := &cleanupQueue{}
	heap.Init(pending)

	logrus.Debug("seeding hidden sound cleanup queue")
	for _, sound := range s.SoundProvider.List() {
		if sound.Hidden {
			heap.Push(pending, cleanupItem{id: sound.Id, expiresAt: sound.CreatedAt.Add(s.HiddenSoundTTL)})
		}
	}

	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	resetTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		if pending.Len() == 0 {
			return
		}
		if d := time.Until((*pending)[0].expiresAt); d > 0 {
			timer.Reset(d)
		} else {
			timer.Reset(0)
		}
	}
	resetTimer()

	logrus.Info("starting sound service worker")
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-s.cleanupCh:
			heap.Push(pending, item)
			resetTimer()
		case <-timer.C:
			logrus.Debug("starting hidden sound cleanup")
			now := time.Now()

			for pending.Len() > 0 && !(*pending)[0].expiresAt.After(now) {
				item := heap.Pop(pending).(cleanupItem)

				sound := s.SoundProvider.Get(item.id)
				if sound == nil || !sound.Hidden {
					continue
				}

				logrus.Infof("deleting \"%s\" expired hidden sound", sound.Id)
				if err := s.SoundProvider.Delete(sound); err != nil {
					logrus.Errorf("error deleting hidden sound: %s", err)
				}
			}
			resetTimer()
		}
	}
}

// enqueueCleanup schedules a hidden sound for expiry-based deletion instead
// of waiting for the next full-list scan. Safe to call before Run starts
// (RegisterRoutes always initializes cleanupCh first); a full queue drops
// the item with a warning rather than blocking the caller.
func (s *Service) enqueueCleanup(item cleanupItem) {
	if s.cleanupCh == nil {
		return
	}

	select {
	case s.cleanupCh <- item:
	default:
		logrus.Warnf("cleanup queue full, sound %q may not be cleaned up until next restart", item.id)
	}
}

func (s *Service) listSound(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	sounds := make([]*Sound, 0)
	for _, sound := range s.SoundProvider.List() {
		if sound.Hidden {
			continue
		}

		sounds = append(sounds, sound)
	}

	sortSounds(sounds, r.URL.Query().Get("sort"), r.URL.Query().Get("order"))

	_ = json.NewEncoder(w).Encode(sounds)
}

// sortSounds sorts in place by sortBy ("name", "play_count", or "created_at");
// an unrecognized sortBy is a no-op, leaving the caller's existing order. order
// of "desc" reverses the comparison; anything else (including empty) is ascending.
func sortSounds(sounds []*Sound, sortBy, order string) {
	var less func(i, j int) bool

	switch sortBy {
	case "name":
		less = func(i, j int) bool { return sounds[i].Name < sounds[j].Name }
	case "play_count":
		less = func(i, j int) bool { return sounds[i].PlayCount < sounds[j].PlayCount }
	case "created_at":
		less = func(i, j int) bool { return sounds[i].CreatedAt.Before(sounds[j].CreatedAt) }
	default:
		return
	}

	if order == "desc" {
		sort.Slice(sounds, func(i, j int) bool { return less(j, i) })
		return
	}

	sort.Slice(sounds, func(i, j int) bool { return less(i, j) })
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
		s.enqueueCleanup(cleanupItem{id: sound.Id, expiresAt: sound.CreatedAt.Add(s.HiddenSoundTTL)})
		break
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sound)
}

func (s *Service) getSound(w http.ResponseWriter, r *http.Request) {
	var sound *Sound

	// load existing values
	sound = s.SoundProvider.Get(mux.Vars(r)["soundId"])
	if sound == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
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
	if !(0 < len(requestSound.Name) && len(requestSound.Name) < 50) {
		service.WriteErrorResponse(w, service.NewNotAcceptableError("requestSound names must be between 1 and 50 characters"))
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

	deletedGroups, err := DeleteSoundWithGroups(s.GroupProvider, s.SoundProvider, sound)
	if err != nil && err != mux.ErrNotFound {
		service.WriteErrorResponse(w, err)
		return
	}

	for _, group := range deletedGroups {
		s.WebsocketService.BroadcastMessage(GroupMessage{
			Type:  websocket.DeleteGroupMessageType,
			Group: group,
		})
	}

	s.WebsocketService.BroadcastMessage(SoundMessage{
		Type:  websocket.DeleteSoundMessageType,
		Sound: sound,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) playSound(w http.ResponseWriter, r *http.Request) {
	_sound := s.SoundProvider.Get(mux.Vars(r)["soundId"])
	if _sound == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.playQueue.EnqueueSounds(*_sound)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) downloadSound(w http.ResponseWriter, r *http.Request) {
	var sound *Sound
	var err error

	sound = s.SoundProvider.Get(mux.Vars(r)["soundId"])
	if sound == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "max-age=31536000")

	err = s.SoundProvider.ReadAudio(sound, w)
	if err != nil {
		service.WriteErrorResponse(w, err)
	}
}

func (s *Service) listGroup(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
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
		if s.SoundProvider.Get(requestGroup.SoundIds[i]) == nil {
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
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&group)
	if err != nil {
		service.WriteErrorResponse(w, err)
		return
	}

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
	if group == nil {
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
		if s.SoundProvider.Get(requestGroup.SoundIds[i]) == nil {
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
		Type:  websocket.UpdateGroupMessageType,
		Group: group,
	})
}

func (s *Service) deleteGroup(w http.ResponseWriter, r *http.Request) {
	var err error
	var group *Group

	group = s.GroupProvider.Get(mux.Vars(r)["groupId"])

	if group != nil {
		err = s.GroupProvider.Delete(group)
	}

	if err != nil && err != mux.ErrNotFound {
		service.WriteErrorResponse(w, err)
		return
	}

	if group != nil {
		s.WebsocketService.BroadcastMessage(GroupMessage{
			Type:  websocket.DeleteGroupMessageType,
			Group: group,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) playGroup(w http.ResponseWriter, r *http.Request) {
	group := s.GroupProvider.Get(mux.Vars(r)["groupId"])
	if group == nil {
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
	sounds := make([]*Sound, 0)
	groups := make([]*Group, 0)

	q := r.URL.Query().Get("q")

	groups = s.GroupProvider.Search(q)

	for _, sound := range s.SoundProvider.Search(q) {
		if sound.Hidden {
			continue
		}

		sounds = append(sounds, sound)
	}

	sortSounds(sounds, r.URL.Query().Get("sort"), r.URL.Query().Get("order"))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"sounds": sounds, "groups": groups})
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
	s.enqueueCleanup(cleanupItem{id: sound.Id, expiresAt: sound.CreatedAt.Add(s.HiddenSoundTTL)})

	// enqueue playback
	s.playQueue.EnqueueSounds(*sound)

	w.WriteHeader(http.StatusAccepted)
}
