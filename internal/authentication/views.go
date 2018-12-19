package authentication

import (
	"encoding/json"
	"github.com/thedevsaddam/govalidator"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"speakerbob/internal"
	"time"
)

const AuthCookieName = "speakerbob"

type UnauthenticatedResponse struct {
	Message string `json:"message"`
}

type BadRequestResponse struct {
	Message string `json:"message"`
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	// short circuit if we are already logged in
	if c, _ := r.Cookie(AuthCookieName); c != nil {
		if internal.GetRedisClient().Exists(getCookieKey(r.RemoteAddr, c)).Val() == 1 {
			w.WriteHeader(http.StatusNoContent)
			_, _ = w.Write([]byte(""))
			return
		}
	}

	// validate data
	var data LoginForm
	e := govalidator.New(govalidator.Options{
		Request: r,
		Data:    &data,
		Rules:   govalidator.MapData{
			"username": []string{"required"},
			"password": []string{"required"},
		},
	}).ValidateJSON()

	// validate form
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

	// validate username
	var user User
	if err := internal.GetDB().Select([]string{"id", "password"}).Where("username = ?", data.Username).First(&user).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"Invalid credentials."})
		return
	}

	// validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BadRequestResponse{"Invalid credentials."})
		return
	}

	// determine cookie expiry
	var cookieExpire time.Time
	if internal.GetConfig().CookieTTL.Seconds() > 0 {
		cookieExpire = time.Now().Add(internal.GetConfig().CookieTTL)
	}

	// create the cookie
	var cookie = &http.Cookie{
		Name:    AuthCookieName,
		Value:   internal.GetUUID(),
		Path:    "/",
		Domain:  r.Host,
		Expires: cookieExpire,
	}
	http.SetCookie(w, cookie)

	// store the cookie value
	internal.GetRedisClient().Set(getCookieKey(r.RemoteAddr, cookie), user.Id, internal.GetConfig().CookieTTL)

	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte(""))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(AuthCookieName)

	// if a cookie is not set then there is no work
	if err != http.ErrNoCookie {
		internal.GetRedisClient().Del(getCookieKey(r.RemoteAddr, cookie))
	}

	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte(""))
}
