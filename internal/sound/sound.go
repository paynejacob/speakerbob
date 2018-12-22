package sound

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/thedevsaddam/govalidator"
	"io"
	"net/http"
	"net/url"
	"os"
	"speakerbob/internal/websocket"
	"strconv"
)

type BadRequestResponse struct {
	Message string `json:"websocket"`
}

type ListResponse struct {
	Count   int     `json:"count"`
	Offset  int     `json:"offset"`
	Results []Sound `json:"results"`
}

type PlaySoundForm struct {
	Channels []string `json:"channels"`
}

type Service struct {
	backend Backend
	pageSize        int
	maxSoundLength  int

	db    *gorm.DB
	wsService    *websocket.Service
}

func NewService(backendURL string, pageSize int, maxSoundLength int, db *gorm.DB, wsService *websocket.Service) *Service {
	var backend Backend
	parsedUrl, err := url.Parse(backendURL)
	if err != nil {
		panic("invalid backend url")
	}

	switch parsedUrl.Scheme {
	case "local":
		parsedURL, err := url.Parse(backendURL)
		if err != nil {
			panic("invalid local backend url")
		}
		backend = NewlocalBackend(parsedURL.Path)
	case "minio://":
		parsedURL, err := url.Parse(backendURL)
		if err != nil {
			panic("invalid minio backend url")
		}

		password, _ := parsedURL.User.Password()
		backend = NewMinioBackend(parsedURL.Host, parsedURL.User.Username(), password, parsedURL.Query()["use_ssl"][0] == "1", parsedURL.Path[1:])
	default:
		panic(fmt.Sprintf("\"%s\" is not a valid sound backend url", parsedUrl.Scheme))
	}

	db.AutoMigrate(&Sound{}, &Macro{}, &PositionalSound{})
	return &Service{backend, pageSize, maxSoundLength, db, wsService}
}

func (s *Service) RegisterRoutes(router *mux.Router, subpath string) {
	router.HandleFunc(fmt.Sprintf("%s/sound", subpath), s.ListSound).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/sound", subpath), s.CreateSound).Methods("POST")
	router.HandleFunc(fmt.Sprintf("%s/sound/{id}", subpath), s.GetSound).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/sound/{id}/download", subpath), s.DownloadSound).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/sound/{id}/play", subpath), s.PlaySound).Methods("POST")

	router.HandleFunc(fmt.Sprintf("%s/macro", subpath), s.ListMacro).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/macro", subpath), s.CreateMacro).Methods("POST")
	router.HandleFunc(fmt.Sprintf("%s/macro/{id}", subpath), s.GetMacro).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/macro/{id}/download", subpath), s.DownloadMacro).Methods("GET")

	router.HandleFunc(fmt.Sprintf("%s/speak", subpath), s.DownloadMacro).Methods("GET")
}

func (s *Service) ListSound(w http.ResponseWriter, r *http.Request) {
	resp := &ListResponse{0, 0, make([]Sound, 0)}

	if offsetStr, ok := r.URL.Query()["offset"]; ok {
		resp.Offset, _ = strconv.Atoi(offsetStr[0])
	}

	s.db.Model(&Sound{}).Where("visible = ?", true).Count(&resp.Count)
	s.db.Limit(s.pageSize).Offset(resp.Offset).Find(&resp.Results)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Service) GetSound(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	sound := &Sound{}

	s.db.Where("id = ?", id).First(&sound)

	w.Header().Set("Content-Type", "application/json")

	if sound.Id == id {
		_ = json.NewEncoder(w).Encode(sound)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Service) CreateSound(w http.ResponseWriter, r *http.Request) {
	var sound Sound

	// Validate the json Payload
	e := govalidator.New(govalidator.Options{
		Request: r,
		Rules: govalidator.MapData{
			"name":       []string{"required"},
			"file:sound": []string{"required"},
		},
	}).Validate()

	// check for form validation errors
	if len(e) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")

		if msg := e.Get("_error"); msg == "unexpected EOF" || msg == "EOF" {
			_ = json.NewEncoder(w).Encode(BadRequestResponse{"Invalid JSON."})
		} else {
			_ = json.NewEncoder(w).Encode(e)
		}
		return
	}

	// create the sound record
	sound = NewSound(r.FormValue("name"), r.FormValue("nsfw") == "true", true)

	// dump the uploaded file to disk
	tmpFilePath := fmt.Sprintf("/tmp/%s", sound.Id)
	outFile, _ := os.Create(tmpFilePath)
	file, _, _ := r.FormFile("sound")
	_, _ = io.Copy(outFile, file)
	_ = file.Close()
	_ = outFile.Close()
	defer func() { _ = os.Remove(tmpFilePath) }()

	// validate the file length
	if length, err := getAudioDuration(tmpFilePath); length > s.maxSoundLength {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(BadRequestResponse{fmt.Sprintf("sound file length may not exceed %d seconds", s.maxSoundLength)})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// normalize the audio file
	normalPath, err := normalizeAudio(tmpFilePath)
	defer func() { _ = os.Remove(normalPath) }()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// upload the audio file to minio
	normalFile, err := os.Open(normalPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	if err := s.backend.PutSound(sound, normalFile); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create the db record
	s.db.Create(&sound)

	// write the response
	_ = json.NewEncoder(w).Encode(sound)
}

func (s *Service) DownloadSound(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var soundRecord Sound

	s.db.Model(&Sound{}).Where("id = ?", id).First(&soundRecord)

	if soundRecord.Id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if s.backend.ServeRedirect() {
		if err := s.backend.RedirectSound(soundRecord, w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	file, err := s.backend.GetSound(soundRecord)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp3", id))
	w.Header().Set("Content-Type", "audio/mpeg")
	_, _ = io.Copy(w, file)
}

func (s *Service) PlaySound(w http.ResponseWriter, r *http.Request)  {
	id := mux.Vars(r)["id"]
	var sound Sound

	if s.db.Select("Id", "NSFW").Where("id = ?", id).First(&sound); sound.Id != "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	channels, _ := r.URL.Query()["channel"]
	if len(channels) == 0 {
		channels = append(channels, "*")
	}

	channelSet := websocket.ChannelSet{}
	for _, channel := range channels {
		channelSet.Add(&websocket.Channel{channel})
	}

	message := websocket.NewPlaySoundMessage(channelSet, sound.Id, sound.NSFW)
	s.wsService.SendMessage(message)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) ListMacro(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) GetMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) CreateMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) DownloadMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) TextToSpeech(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
