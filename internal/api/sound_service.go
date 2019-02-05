package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/thedevsaddam/govalidator"
	"github.com/watson-developer-cloud/go-sdk/texttospeechv1"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type ListSoundResponse struct {
	Count   int     `json:"count"`
	Offset  int     `json:"offset"`
	Results []Sound `json:"results"`
}

type ListMacroResponse struct {
	Count   int     `json:"count"`
	Offset  int     `json:"offset"`
	Results []Macro `json:"results"`
}

type DetailMacroResponse struct {
	Macro

	Sounds []string `json:"sounds"`
}

type SpeakForm struct {
	Text     string   `json:"text"`
	NSFW     bool     `json:"nsfw"`
	Channels []string `json:"channels"`
}

type MacroForm struct {
	Name   string   `json:"name"`
	Sounds []string `json:"sounds"`
}

type SoundService struct {
	backend        SoundBackend
	pageSize       int
	maxSoundLength int

	db            *gorm.DB
	wsService     *WebsocketService
	searchService *SearchService
	bluemixKey    string
}

func NewSoundService(backendURL string, pageSize int, maxSoundLength int, db *gorm.DB, wsService *WebsocketService, searchService *SearchService, bluemixKey string) *SoundService {
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
	return &SoundService{backend, pageSize, maxSoundLength, db, wsService, searchService, bluemixKey}
}

func (s *SoundService) RegisterRoutes(parent *mux.Router, prefix string) *mux.Router {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/sound", s.ListSound).Methods("GET")
	router.HandleFunc("/sound", s.CreateSound).Methods("POST")
	router.HandleFunc("/sound/{id}", s.GetSound).Methods("GET")
	router.HandleFunc("/sound/{id}/download", s.DownloadSound).Methods("GET")
	router.HandleFunc("/sound/{id}/play", s.PlaySound).Methods("POST")

	router.HandleFunc("/macro", s.ListMacro).Methods("GET")
	router.HandleFunc("/macro", s.CreateMacro).Methods("POST")
	router.HandleFunc("/macro/{id}", s.GetMacro).Methods("GET")
	router.HandleFunc("/macro/{id}/play", s.PlayMacro).Methods("GET")

	router.HandleFunc("/speak", s.Speak).Methods("POST")

	return router
}

func (s *SoundService) ListSound(w http.ResponseWriter, r *http.Request) {
	resp := &ListSoundResponse{0, 0, make([]Sound, 0)}

	if offsetStr, ok := r.URL.Query()["offset"]; ok {
		resp.Offset, _ = strconv.Atoi(offsetStr[0])
	}

	s.db.Model(&Sound{}).Where("visible = true").Count(&resp.Count)
	s.db.Where("visible = true").Limit(s.pageSize).Offset(resp.Offset).Find(&resp.Results)

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		if msg := e.Get("_error"); msg == "unexpected EOF" || msg == "EOF" {
			_ = json.NewEncoder(w).Encode(MessageResponse{"Invalid JSON."})
		} else {
			_ = json.NewEncoder(w).Encode(e)
		}
		return
	}

	var duplicateName int
	if _ = s.db.Model(&Sound{}).Where("name = ?", r.FormValue("name")).Count(&duplicateName).Limit(1); duplicateName > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(MessageResponse{"the sound name must be unique"})
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
	_ = s.searchService.UpdateResult(sound)

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

	message := NewPlaySoundMessage(channelSet, sound)
	s.wsService.SendMessage(message)

	w.WriteHeader(http.StatusNoContent)
}

func (s *SoundService) ListMacro(w http.ResponseWriter, r *http.Request) {
	resp := &ListMacroResponse{0, 0, make([]Macro, 0)}

	if offsetStr, ok := r.URL.Query()["offset"]; ok {
		resp.Offset, _ = strconv.Atoi(offsetStr[0])
	}

	s.db.Model(&Macro{}).Count(&resp.Count)
	s.db.Limit(s.pageSize).Offset(resp.Offset).Find(&resp.Results)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *SoundService) GetMacro(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	macro := &Macro{}

	s.db.Where("id = ?", id).First(&macro)

	w.Header().Set("Content-Type", "application/json")

	if macro.Id == id {
		_ = json.NewEncoder(w).Encode(macro)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *SoundService) CreateMacro(w http.ResponseWriter, r *http.Request) {

	// run validation
	var data MacroForm
	e := govalidator.New(govalidator.Options{
		Request: r,
		Data:    &data,
		Rules: govalidator.MapData{
			"name":   []string{"required"},
			"sounds": []string{"required"},
		},
	}).ValidateJSON()

	// display form errors
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

	// ensure we have at least 2 songs
	if len(data.Sounds) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MessageResponse{"\"sounds\" must contain at least 2 sounds"})
	}

	// find songs matching ids
	var sounds []Sound
	if err := s.db.Select("id").Model(&Sound{}).Where("id in ?", data.Sounds); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// map songs to ids
	soundMap := make(map[string]Sound, 0)
	for _, sound := range sounds {
		soundMap[sound.Id] = sound
	}

	// ensure a sound was returned for each unique sound id
	if len(sounds) != len(data.Sounds) {
		for _, soundId := range data.Sounds {
			if _, ok := soundMap[soundId]; !ok {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(MessageResponse{fmt.Sprintf("\"%s\" is not a valid source id", soundId)})
				return
			}
		}
	}

	tx := s.db.Begin()

	// create macro
	macro := NewMacro(data.Name)
	if err := tx.Create(&macro).Error; err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create positional sounds
	var positionalSounds []*PositionalSound
	for pos, sound := range data.Sounds {
		nps := NewPositionalSound(pos, soundMap[sound], *macro)

		positionalSounds = append(positionalSounds, nps)
		if err := tx.Create(&nps).Error; err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// commit transaction
	if tx.Commit().Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	detailMacro := DetailMacroResponse{*macro, data.Sounds}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(detailMacro)

}

func (s *SoundService) PlayMacro(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	channels, _ := r.URL.Query()["channel"]
	var exists bool = false

	const Q = "select s.id, s.NSFW from positional_sound as ps join sounds s on s.id = ps.sound_id where ps.macro_id = \"?\""

	if len(channels) == 0 {
		channels = append(channels, "*")
	}

	channelSet := ChannelSet{}
	for _, channel := range channels {
		channelSet.Add(&Channel{Value: channel})
	}

	sounds, _ := db.Raw(Q, id).Rows()
	defer func() { _ = sounds.Close() }()
	for sounds.Next() {
		exists = true
		var sound api.Sound

		_ = db.ScanRows(sounds, &sound)
		s.wsService.SendMessage(NewPlaySoundMessage(channels, sound))
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *SoundService) Speak(w http.ResponseWriter, r *http.Request) {
	if s.bluemixKey == "" {
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
		s.wsService.SendMessage(NewPlaySoundMessage(channelSet, sound))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(sound)
		return
	} else {
		sound = NewSound(hashedName, data.NSFW, false)
	}

	textToSpeech, err := texttospeechv1.NewTextToSpeechV1(&texttospeechv1.TextToSpeechV1Options{
		IAMApiKey: s.bluemixKey,
		URL:       "https://gateway-wdc.watsonplatform.net/text-to-speech/api",
	})

	format := "audio/wav"

	resp, err := textToSpeech.Synthesize(&texttospeechv1.SynthesizeOptions{
		Text:   &data.Text,
		Accept: &format,
	})

	// validate audio response
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// put sound in backend
	if err := s.backend.PutSound(sound, textToSpeech.GetSynthesizeResult(resp)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create sound and send play message
	s.db.Create(&sound)
	sound.Visible = false
	s.db.Save(&sound)
	s.wsService.SendMessage(NewPlaySoundMessage(channelSet, sound))

	// return response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sound)
}
