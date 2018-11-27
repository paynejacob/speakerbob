package services

import (
	"net/http"
)

//GetSpeak: Plays the given text on all clients
func GetSpeak(w http.ResponseWriter, r *http.Request) {
	// validate payload
	// request wave file or load existing
	// create new sound if not exist
	// upload to minio if not exist
	// send play message
	// return 200
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(""))
}
