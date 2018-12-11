package services

import (
	"encoding/json"
	"net/http"
	"speakerbob/internal"
)

const AuthCookieName = "speakerbob"


type UnauthenticatedResponse struct {
	Message string `json:"message"`
}


func PostLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(""))
}

func GetLogout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(""))
}

func AuthenticationMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cookie, err := r.Cookie(AuthCookieName)

	// Fail if no cookie is found or the cookie value does not exist in redis
	if err == http.ErrNoCookie || internal.GetRedisClient().Exists(cookie.Value).Val() == 0 {
		var resp UnauthenticatedResponse = UnauthenticatedResponse{"You must be logged in to preform this action"}
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(resp)
	}
	
	next(rw, r)
}