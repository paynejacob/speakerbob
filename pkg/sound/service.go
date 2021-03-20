package sound

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Service struct {
	soundProvider *Provider
}

const cleanupInterval = 6 * time.Hour
const soundCreateGracePeriod = 10 * time.Minute

func NewService(soundStore *Provider) *Service {
	return &Service{soundProvider: soundStore}
}

func (s *Service) RegisterRoutes(parent *mux.Router, prefix string) {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/", s.list).Methods("GET")
	router.HandleFunc("/", s.create).Methods("POST")
	router.HandleFunc("/{soundId}/", s.update).Methods("PATCH")
	router.HandleFunc("/{soundId}/", s.delete).Methods("DELETE")
	router.HandleFunc("/{soundId}/download/", s.download).Methods("GET")
}

func (s *Service) Run() {
	var sounds []Sound
	var sound Sound
	var err error
	var now time.Time

	logrus.Info("starting sound service worker")

	if err = s.soundProvider.HydrateSearch(); err != nil {
		logrus.Panicf("Error hydrating search index! %d", err)
	}

	for range time.Tick(cleanupInterval) {
		logrus.Debug("starting database garbage collection")
		if err = s.soundProvider.db.RunValueLogGC(0.5); badger.ErrNoRewrite != err {
			logrus.Errorf("error durring database garbage collection: %d", err)
		}

		logrus.Debug("starting uninitialized sound cleanup")
		now = time.Now()
		sounds, err = s.soundProvider.AllSounds()
		if err != nil {
			logrus.Errorf("error listing uninitalized sounds: %d", err)
		}
		for i := range sounds {
			if sounds[i].Name == "" && now.Sub(sounds[i].CreatedAt) > soundCreateGracePeriod {
				logrus.Infof("deleting \"%s\" expired uninitialized sound", sounds[i].Id)

				err = s.soundProvider.DeleteSound(sound)
				if err != nil {
					logrus.Errorf("error deleting uninitalized sound: %d", err)
				}
			}
		}
	}
}

func (s *Service) create(w http.ResponseWriter, r *http.Request) {
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

		sound, err = s.soundProvider.CreateSound(fileHeader.Filename, data)
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

func (s *Service) update(w http.ResponseWriter, r *http.Request) {
	var currentSound Sound
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
	currentSound.Id = soundId
	err = s.soundProvider.GetSound(&currentSound)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// write user changes
	currentSound.Name = sound.Name
	currentSound.NSFW = sound.NSFW
	err = s.soundProvider.SaveSound(currentSound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) delete(w http.ResponseWriter, r *http.Request) {
	var sound Sound

	sound.Id = mux.Vars(r)["soundId"]

	err := s.soundProvider.DeleteSound(sound)

	if err != nil && err != mux.ErrNotFound {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) {
	var err error
	var sounds []Sound

	q := r.URL.Query().Get("q")

	sounds, err = s.soundProvider.SearchSounds(graph.Tokenize(q))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(sounds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) download(w http.ResponseWriter, r *http.Request) {
	var sound Sound
	var err error

	sound.Id = mux.Vars(r)["soundId"]

	w.Header().Set("Content-Type", "audio/mp3")
	w.Header().Set("Cache-Control", "max-age=2592000s")

	err = s.soundProvider.GetSoundAudio(sound, w)

	if err == badger.ErrKeyNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
