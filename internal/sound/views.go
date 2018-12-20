package sound

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go"
	"github.com/thedevsaddam/govalidator"
	"io"
	"net/http"
	"os"
	"speakerbob/internal"
	"strconv"
)

type BadRequestResponse struct {
	Message string `json:"message"`
}

type ListSoundResponse struct {
	Count int `json:"count"`
	Offset int `json:"offset"`
	Results []Sound `json:"results"`
}

func init()  {
	govalidator.AddCustomRule("uniqueSoundName", func(field string, rule string, message string, value interface{}) error {
		count := 0
		internal.GetDB().Model(&Sound{}).Where("name = ?", value).Count(&count)
		if count > 0 {
			if message != "" {
				return errors.New(message)
			}
			return fmt.Errorf("sound with the name \"%s\" already exists", value)
		}

		return nil
	})
}

func ListSound(w http.ResponseWriter, r *http.Request) {
	resp := &ListSoundResponse{0, 0, make([]Sound, 0)}

	if offsetStr, ok := r.URL.Query()["offset"]; ok {
		resp.Offset, _ = strconv.Atoi(offsetStr[0])
	}

	internal.GetDB().Model(&Sound{}).Where("visible = ?", true).Count(&resp.Count)
	internal.GetDB().Limit(internal.GetConfig().PageSize).Offset(resp.Offset).Find(&resp.Results)


	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func GetSound(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	sound := &Sound{}

	internal.GetDB().Where("id = ?", id).First(&sound)


	w.Header().Set("Content-Type", "application/json")

	if sound.Id == id {
		_ = json.NewEncoder(w).Encode(sound)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func CreateSound(w http.ResponseWriter, r *http.Request) {
	var sound Sound

	// Validate the json Payload
	e := govalidator.New(govalidator.Options{
		Request: r,
		Rules:   govalidator.MapData{
			"name": []string{"required", "uniqueSoundName"},
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
	defer func() {_ = os.Remove(tmpFilePath)}()

	// validate the file length
	if length, err := getAudioDuration(tmpFilePath); length > internal.GetConfig().MaxSoundLength {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(BadRequestResponse{fmt.Sprintf("sound file length may not exceed %d seconds", internal.GetConfig().MaxSoundLength)})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"an error occurred processing the audio file"})
		return
	}

	// normalize the audio file
	normalPath, err := normalizeAudio(tmpFilePath)
	defer func() {_ = os.Remove(normalPath)}()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"an error occurred processing the audio file"})
		return
	}

	// upload the audio file to minio
	if _, err := internal.GetMinioClient().FPutObject(internal.GetConfig().SoundBucketName, sound.Id, normalPath, minio.PutObjectOptions{}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"an error occurred processing the audio file"})
		return
	}

	// create the db record
	internal.GetDB().Create(&sound)

	// write the response
	_ = json.NewEncoder(w).Encode(sound)
}

func DownloadSound(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var exists int

	internal.GetDB().Model(&Sound{}).Where("id = ?", id).Count(&exists).Limit(1)

	if exists != 1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	obj, err := internal.GetMinioClient().GetObject(internal.GetConfig().SoundBucketName, id, minio.GetObjectOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	objStats, _ := obj.Stat()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp3", id))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", objStats.Size))
	_, _ = io.Copy(w, obj)
}

type ListMacroResponse struct {
	Count int `json:"count"`
	Offset int `json:"offset"`
	Results []Macro `json:"results"`
}

func ListMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(""))
}

func GetMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(""))
}

func CreateMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(""))
}

func DownloadMacro(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(""))
}
