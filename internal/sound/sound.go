package sound

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/minio/minio-go"
	"github.com/thedevsaddam/govalidator"
	"io"
	"net/http"
	"os"
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

// TODO storage backend

func NewService(soundBucketName string, pageSize int, maxSoundLength int, db *gorm.DB, minio *minio.Client) *Service {
	ensureBucket(soundBucketName, minio)
	db.AutoMigrate(&Sound{}, &Macro{}, &PositionalSound{})
	return &Service{soundBucketName, pageSize, maxSoundLength, db, minio}
}

type Service struct {
	soundBucketName string
	pageSize        int
	maxSoundLength  int

	db    *gorm.DB
	minio *minio.Client
}

func (s *Service) RegisterRoutes(router *mux.Router, subpath string) {
	router.HandleFunc(fmt.Sprintf("%s/sound", subpath), s.ListSound).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/sound", subpath), s.CreateSound).Methods("POST")
	router.HandleFunc(fmt.Sprintf("%s/sound/{id}", subpath), s.GetSound).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/sound/{id}/download", subpath), s.DownloadSound).Methods("GET")

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
			"name":       []string{"required", "uniqueSoundName"}, // TODO unique name
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
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"an error occurred processing the audio file"})
		return
	}

	// normalize the audio file
	normalPath, err := normalizeAudio(tmpFilePath)
	defer func() { _ = os.Remove(normalPath) }()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"an error occurred processing the audio file"})
		return
	}

	// upload the audio file to minio
	if _, err := s.minio.FPutObject(s.soundBucketName, sound.Id, normalPath, minio.PutObjectOptions{}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"an error occurred processing the audio file"})
		return
	}

	// create the db record
	s.db.Create(&sound)

	// write the response
	_ = json.NewEncoder(w).Encode(sound)
}

func (s *Service) DownloadSound(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var exists int

	s.db.Model(&Sound{}).Where("id = ?", id).Count(&exists).Limit(1)

	if exists != 1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	obj, err := s.minio.GetObject(s.soundBucketName, id, minio.GetObjectOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	objStats, _ := obj.Stat()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp3", id))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", objStats.Size))
	_, _ = io.Copy(w, obj)
}

// TODO play
// TODO search
// TODO this
func (s *Service) ListMacro(w http.ResponseWriter, r *http.Request) {
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
