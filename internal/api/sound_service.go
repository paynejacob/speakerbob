package api

import (
	"encoding/json"
	"fmt"
	bluemix "github.com/IBM-Cloud/bluemix-go/session"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/thedevsaddam/govalidator"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type ListSoundResponse struct {
	Count int `json:"count"`
	Offset int `json:"offset"`
	Results []Sound `json:"results"`
}

type SpeakForm struct {
	Text     string   `json:"text"`
	NSFW     bool     `json:"nsfw"`
	Channels []string `json:"channels"`
}

type SearchResult Sound

func (SearchResult) Type() string {
	return "sound"
}

func (r SearchResult) Key() string {
	return r.Id
}

func (r SearchResult) IndexValue() string {
	return r.Name
}

func (r SearchResult) Object() interface{} {
	return r
}

type SoundService struct {
	backend        SoundBackend
	pageSize       int
	maxSoundLength int

	db             *gorm.DB
	wsService      *WebsocketService
	searchService  *SearchService
	blueMixSession *bluemix.Session
}

func NewSoundService(backendURL string, pageSize int, maxSoundLength int, db *gorm.DB, wsService *WebsocketService, searchService *SearchService, blueMixSession *bluemix.Session) *SoundService {
	var backend SoundBackend
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
		backend = NewSoundLocalBackend(parsedURL.Path)
	case "minio://":
		parsedURL, err := url.Parse(backendURL)
		if err != nil {
			panic("invalid minio backend url")
		}

		password, _ := parsedURL.User.Password()
		backend = NewSoundMinioBackend(parsedURL.Host, parsedURL.User.Username(), password, parsedURL.Query()["use_ssl"][0] == "1", parsedURL.Path[1:])
	default:
		panic(fmt.Sprintf("\"%s\" is not a valid sound backend url", parsedUrl.Scheme))
	}

	db.AutoMigrate(&Sound{}, &Macro{}, &PositionalSound{})
	return &SoundService{backend, pageSize, maxSoundLength, db, wsService, searchService, blueMixSession}
}

func (s *SoundService) RegisterRoutes(parent *mux.Router, prefix string) *mux.Router {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("", s.ListSound).Methods("GET")
	router.HandleFunc("", s.CreateSound).Methods("POST")
	router.HandleFunc("/{id}", s.GetSound).Methods("GET")
	router.HandleFunc("/{id}/download", s.DownloadSound).Methods("GET")
	router.HandleFunc("/{id}/play", s.PlaySound).Methods("POST")

	router.HandleFunc("/", s.ListMacro).Methods("GET")
	router.HandleFunc("/", s.CreateMacro).Methods("POST")
	router.HandleFunc("/{id}", s.GetMacro).Methods("GET")
	router.HandleFunc("/{id}/download", s.DownloadMacro).Methods("GET")

	router.HandleFunc("/speak", s.Speak).Methods("POST")

	return router
}

func (s *SoundService) ListSound(w http.ResponseWriter, r *http.Request) {
	resp := &ListSoundResponse{0, 0, make([]Sound, 0)}

	if offsetStr, ok := r.URL.Query()["offset"]; ok {
		resp.Offset, _ = strconv.Atoi(offsetStr[0])
	}

	s.db.Model(&Sound{}).Where("visible = ?", true).Count(&resp.Count)
	s.db.Limit(s.pageSize).Offset(resp.Offset).Find(&resp.Results)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *SoundService) GetSound(w http.ResponseWriter, r *http.Request) {
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

func (s *SoundService) CreateSound(w http.ResponseWriter, r *http.Request) {
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
			_ = json.NewEncoder(w).Encode(MessageResponse{"Invalid JSON."})
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
		_ = json.NewEncoder(w).Encode(MessageResponse{fmt.Sprintf("sound file length may not exceed %d seconds", s.maxSoundLength)})
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

	// update the search index
	_ = s.searchService.UpdateResult(SearchResult(sound))

	// write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sound)
}

func (s *SoundService) DownloadSound(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.wav", id))
	w.Header().Set("Content-Type", "audio/wav")
	_, _ = io.Copy(w, file)
}

func (s *SoundService) PlaySound(w http.ResponseWriter, r *http.Request) {
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

	channelSet := ChannelSet{}
	for _, channel := range channels {
		channelSet.Add(&Channel{Value: channel})
	}

	message := NewPlaySoundMessage(channelSet, sound.Id, sound.NSFW)
	s.wsService.SendMessage(message)

	w.WriteHeader(http.StatusNoContent)
}

func (s *SoundService) ListMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *SoundService) GetMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *SoundService) CreateMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *SoundService) DownloadMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *SoundService) Speak(w http.ResponseWriter, r *http.Request) {
	if s.blueMixSession == nil {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	// validate form data
	var data SpeakForm
	e := govalidator.New(govalidator.Options{
		Request: r,
		Data:    &data,
		Rules: govalidator.MapData{
			"text":     []string{"required"},
			"nsfw":     []string{},
			"channels": []string{},
		},
	}).ValidateJSON()

	if len(e) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")

		if msg := e.Get("_error"); msg == "unexpected EOF" || msg == "EOF" {
			_ = json.NewEncoder(w).Encode(MessageResponse{"Invalid JSON."})
		} else {
			_ = json.NewEncoder(w).Encode(e)
		}
		return
	}

	// create channelSet
	channelSet := ChannelSet{}
	if len(data.Channels) == 0 {
		channelSet.Add(&Channel{Value: "*"})
	} else {
		for _, channel := range data.Channels {
			channelSet.Add(&Channel{Value: channel})
		}
	}

	// check if sound exists and return it if so
	sound := Sound{}
	hashedName := hashSpeakName(data.Text)
	s.db.Where("name = ?", hashedName).First(&sound)
	if sound.Id != "" {
		s.wsService.SendMessage(NewPlaySoundMessage(channelSet, sound.Id, sound.NSFW))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(sound)
		return
	} else {
		sound = NewSound(hashedName, data.NSFW, false)
	}

	// create audio
	resp, err := s.blueMixSession.Config.HTTPClient.PostForm(
		"https://gateway-wdc.watsonplatform.net/text-to-speech/api/v1/synthesize",
		url.Values{
			"accept": []string{"audio/wav"},
			"text":   []string{data.Text},
		})

	// validate audio response
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// put sound in backend
	if err := s.backend.PutSound(sound, resp.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create sound and send play message
	s.db.Create(&sound)
	s.wsService.SendMessage(NewPlaySoundMessage(channelSet, sound.Id, sound.NSFW))

	// return response
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sound)
}
