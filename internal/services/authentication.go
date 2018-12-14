package services

import (
	"encoding/json"
	"net/http"
	"speakerbob/internal"
	"time"
	"github.com/google/uuid"
)

const AuthCookieName = "speakerbob"


type UnauthenticatedResponse struct {
	Message string `json:"message"`
}

type LoginForm struct {
	Username string `json:"username"`
	Password string	`json:"password"`
}


func Login(w http.ResponseWriter, r *http.Request) {

	// validate data
	var decoder = json.NewDecoder(r.Body)
	var err = decoder.Decode(&LoginForm{})

	if err != nil {
		
	}






	// create the cookie
	var cookie = &http.Cookie{
		Name:       AuthCookieName,
		Value:      uuid.New().String(),
		Path:       "/",
		Domain:     r.Host,
		Expires:    time.Now().Add(internal.GetConfig().CookieTTL),
	}
	http.SetCookie(w, cookie)

	// TODO: values in redis

	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte(""))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(AuthCookieName)

	// if a cookie is not set then there is no work
	if err != http.ErrNoCookie {
		internal.GetRedisClient().Del(cookie.Value)
	}

	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte(""))
}

func AuthenticationMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cookie, err := r.Cookie(AuthCookieName)

	// Fail if no cookie is found or the cookie value does not exist in redis
	if err == http.ErrNoCookie || internal.GetRedisClient().Exists(cookie.Value).Val() == 0 {
		var resp = UnauthenticatedResponse{"You must be authenticated to preform this action"}
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(resp)
	}
	
	next(rw, r)
}