package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/thedevsaddam/govalidator"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"time"
)

var backendProtoRegex = regexp.MustCompile("(.*)://.*?")

type UnauthenticatedResponse struct {
	Message string `json:"websocket"`
}

type BadRequestResponse struct {
	Message string `json:"message"`
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewService(backendURL string, cookieName string, ttl time.Duration, db *gorm.DB) *Service {
	var backend Backend

	if matches := backendProtoRegex.FindStringSubmatch(backendURL); len(matches) > 0 {
		proto := matches[0]
		switch proto {
		case "memory://":
			backend = NewMemoryBackend(ttl)
		default:
			panic(fmt.Sprintf("\"%s\" is not a valid authentication backend proto", proto))
		}
	}

	db.AutoMigrate(&User{})

	return &Service{backend, cookieName, db}
}

type Service struct {
	backend    Backend
	cookieName string

	db *gorm.DB
}

func (s *Service) RegisterRoutes(router *mux.Router, subpath string) {
	router.HandleFunc(fmt.Sprintf("%s/login", subpath), s.Login).Methods("POST")
	router.HandleFunc(fmt.Sprintf("%s/logout", subpath), s.Logout).Methods("GET")
}

func (s *Service) AuthenticationMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cookie, err := r.Cookie(s.cookieName)

	// Fail if no cookie is found or the cookie value does not exist in redis
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	userId, _ := s.backend.UserId(r.RemoteAddr, cookie.Value)
	if userId == "" {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	ctx := context.WithValue(r.Context(), "userId", userId)

	next(rw, r.WithContext(ctx))
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	// short circuit if we are already logged in
	if c, _ := r.Cookie(s.cookieName); c != nil {
		if _, err := s.backend.UserId(r.RemoteAddr, c.Value); err == nil {
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
		Rules: govalidator.MapData{
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
	if err := s.db.Select([]string{"id", "password"}).Where("username = ?", data.Username).First(&user).Error; err != nil {
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
	if s.backend.TTL().Seconds() > 0 {
		cookieExpire = time.Now().Add(s.backend.TTL())
	}

	token, err := s.backend.NewToken(r.RemoteAddr, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create the cookie
	var cookie = &http.Cookie{
		Name:    s.cookieName,
		Value:   token,
		Path:    "/",
		Domain:  r.Host,
		Expires: cookieExpire,
	}
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte(""))
}

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {
	// if a cookie is not set then there is no work
	if cookie, err := r.Cookie(s.cookieName); err != http.ErrNoCookie {
		s.backend.InvalidateToken(r.RemoteAddr, cookie.Value)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) ListUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) GetUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
