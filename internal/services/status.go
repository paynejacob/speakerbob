package services

import (
	"net/http"
)

//GetStatus: Provides inexpensive endpoint for checking if the server is running
func GetStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte(""))
}
