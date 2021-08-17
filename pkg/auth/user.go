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

//go:generate go run github.com/paynejacob/speakerbob/codegen github.com/paynejacob/speakerbob/pkg/auth.User
type User struct {
	Id        string    `json:"id" store:"key"`
	CreatedAt time.Time `json:"created_at"`

	Name       string      `json:"name,omitempty" store:"searchable"`
	Email      string      `json:"-" store:"lookup"`
	Principals []Principal `json:"-" store:"lookup"`
}

func NewUser() User {
	return User{
		Id:        strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt: time.Now(),
	}
}
