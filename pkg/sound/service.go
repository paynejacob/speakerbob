package sound

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

type Service struct {
	soundStore  *Store
	searchIndex *Index
}

func NewService(soundStore *Store) *Service {
	return &Service{soundStore: soundStore, searchIndex: NewIndex()}
}

func (s *Service) RegisterRoutes(parent *mux.Router, prefix string) {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/", s.list).Methods("GET")
	router.HandleFunc("/", s.create).Methods("POST")
	router.HandleFunc("/{soundId}/", s.update).Methods("PATCH")
	router.HandleFunc("/{soundId}/download/", s.download).Methods("GET")
}

func (s *Service) Run() {
	var expiredSounds []Sound
	var sound Sound
	var err error
	var now time.Time

	logrus.Info("starting sound service worker")

	logrus.Info("hydrating sound search index")
	_ = s.soundStore.InitializeCache()
	for _, sound := range s.soundStore.All() {
		s.searchIndex.IndexSound(sound)
	}

	for {
		logrus.Debug("starting uninitialized sound cleanup")
		now = time.Now()
		for _, sound = range s.soundStore.UninitializedSounds() {
			if now.Sub(sound.CreatedAt) > s.soundStore.soundCreateGracePeriod {
				expiredSounds = append(expiredSounds, sound)
			}
		}

		logrus.Debugf("deleting %d expired uninitialized sounds", len(expiredSounds))

		for _, sound = range expiredSounds {
			err = s.soundStore.Delete(sound)
			if err != nil {
				logrus.Errorf("error deleting expired uninitialized sound: %s", err)
			}
		}

		expiredSounds = make([]Sound, 0)

		time.Sleep(s.soundStore.soundCreateGracePeriod)
	}
}

func (s *Service) create(w http.ResponseWriter, r *http.Request) {
	var sound Sound
	var err error

	err = r.ParseMultipartForm(10 << 20)
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

		sound, err = s.soundStore.Create(fileHeader.Filename, data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // TODO: determine file is bad or upload is bad
			return
		}
		break
	}

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
	currentSound, err = s.soundStore.Get(soundId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// write user changes
	currentSound.Name = sound.Name
	currentSound.NSFW = sound.NSFW
	err = s.soundStore.Save(currentSound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// update search index
	s.searchIndex.Clear()
	for _, sound := range s.soundStore.All() {
		s.searchIndex.IndexSound(sound)
	}
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) {
	var err error
	var soundId string
	var sound Sound
	var sounds []Sound

	q := r.URL.Query().Get("q")

	if q == "" {
		sounds = s.soundStore.All()
	} else {
		for _, soundId = range s.searchIndex.Search(q) {
			sound, err = s.soundStore.Get(soundId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			sounds = append(sounds, sound)
		}
	}

	err = json.NewEncoder(w).Encode(sounds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) download(w http.ResponseWriter, r *http.Request) {
	var sound Sound
	var err error
	var _url *url.URL

	soundId := mux.Vars(r)["soundId"]

	sound, err = s.soundStore.Get(soundId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_url, err = s.soundStore.GetDownloadUrl(sound)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, _url.String(), http.StatusTemporaryRedirect)
}
