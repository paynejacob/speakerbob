package services

import (
	"net/http"
	"speakerbob/internal"
)

//GetStatus: Provides inexpensive endpoint for checking if the server is running
//noinspection ALL
func Status(w http.ResponseWriter, r *http.Request) {
	_ = internal.GetRedisClient().Ping()
	_ = internal.GetDB().DB().Ping()

	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte(""))
}
