package auth

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type Principal string

func NewPrincipal(providerName, userId string) Principal {
	return Principal(providerName + "://" + userId)
}

//go:generate go run github.com/paynejacob/hotcereal providergen github.com/paynejacob/speakerbob/pkg/auth.User
type User struct {
	Id        string    `json:"id" hotcereal:"key"`
	CreatedAt time.Time `json:"created_at"`

	Email      string      `json:"-" hotcereal:"lookup"`
	Principals []Principal `json:"-" hotcereal:"lookup"`

	Preferences map[string]string `json:"preferences"`
}

func NewUser() User {
	return User{
		Id:          strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt:   time.Now(),
		Preferences: make(map[string]string, 0),
	}
}
