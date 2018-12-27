package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/thedevsaddam/govalidator"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ListUserResponse struct {
	Count int
	Offset int
	Results []User
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserForm struct {
	DisplayName string `json:"display_name"`
	Password string `json:"password"`
}

type AuthenticationService struct {
	backend    AuthenticationBackend
	cookieName string

	PageSize int

	db *gorm.DB
}

func NewAuthenticationService(backendURL string, cookieName string, ttl time.Duration, pageSize int, db *gorm.DB) *AuthenticationService {
	var backend AuthenticationBackend
	parsedUrl, err := url.Parse(backendURL)
	if err != nil {
		panic("invalid backend url")
	}

	switch parsedUrl.Scheme {
	case "memory":
		backend = NewAuthenticationMemoryBackend(ttl)
	case "":
		backend = NewAuthenticationNoopBackend(ttl)
	default:
		panic(fmt.Sprintf("\"%s\" is not a valid authentication backend proto", parsedUrl.Scheme))
	}

	db.AutoMigrate(&User{})

	return &AuthenticationService{backend, cookieName, pageSize, db}
}

func (s *AuthenticationService) RegisterRoutes(parent *mux.Router, prefix string) *mux.Router {
	router := parent.PathPrefix(prefix).Subrouter()
	userRouter := router.PathPrefix("/user").Subrouter()

	router.HandleFunc("/login", s.Login).Methods("POST")
	router.HandleFunc("/logout", s.Logout).Methods("GET")

	userRouter.Use(s.AuthenticationMiddleware)
	userRouter.HandleFunc("", s.ListUser).Methods("GET")
	userRouter.HandleFunc("/{id}", s.GetUser).Methods("GET")
	userRouter.HandleFunc("/{id}", s.UpdateUser).Methods("PATCH")

	return router
}

func (s *AuthenticationService) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		cookie, err := r.Cookie(s.cookieName)

		// Fail if no cookie is found or the cookie value does not exist in redis
		if err != nil {
			token = ""
		} else {
			token = cookie.Value
		}


		if _, err := s.backend.UserId(r.RemoteAddr, token); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *AuthenticationService) Login(w http.ResponseWriter, r *http.Request) {
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
			_ = json.NewEncoder(w).Encode(MessageResponse{"Invalid JSON."})
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
		_ = json.NewEncoder(w).Encode(MessageResponse{"Invalid credentials."})
		return
	}

	// validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MessageResponse{"Invalid credentials."})
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

func (s *AuthenticationService) Logout(w http.ResponseWriter, r *http.Request) {
	// if a cookie is not set then there is no work
	if cookie, err := r.Cookie(s.cookieName); err != http.ErrNoCookie {
		s.backend.InvalidateToken(r.RemoteAddr, cookie.Value)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *AuthenticationService) ListUser(w http.ResponseWriter, r *http.Request) {
	resp := ListUserResponse{0, 0, make([]User, 0)}

	if offsetStr, ok := r.URL.Query()["offset"]; ok {
		resp.Offset, _ = strconv.Atoi(offsetStr[0])
	}

	s.db.Model(&User{}).Count(&resp.Count)
	s.db.Limit(s.PageSize).Offset(resp.Offset).Find(&resp.Results)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *AuthenticationService) GetUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	user := &User{}

	s.db.Where("id = ?", id).First(&user)

	w.Header().Set("Content-Type", "application/json")

	if user.Id == id {
		_ = json.NewEncoder(w).Encode(user)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *AuthenticationService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	user := &User{}

	if id != r.Context().Value("userId") {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	s.db.Where("id = ?", id).First(&user)

	// validate data
	var data UserForm
	e := govalidator.New(govalidator.Options{
		Request: r,
		Data:    &data,
		Rules: govalidator.MapData{
			"display_name": []string{},
			"password": []string{},
		},
	}).ValidateJSON()

	// validate form
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

	if data.DisplayName != "" {
		user.DisplayName = data.DisplayName
	}

	if data.Password != "" {
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		user.Password = string(passwordHash)
	}

	s.db.Save(&user)
	w.Header().Set("Content-Type", "application/json")
	if user.Id == id {
		_ = json.NewEncoder(w).Encode(user)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
