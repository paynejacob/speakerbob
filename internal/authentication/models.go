package authentication

import (
	"golang.org/x/crypto/bcrypt"
	"speakerbob/internal"
	"time"
)

type User struct {
	Id        string    `gorm:"primary_key;unique;index" json:"id"`
	CreatedAt time.Time `json:"create_at"`

	Username string `gorm:"unique;index"`
	Password string

	DisplayName string
}

func NewUser(username string, password string, displayName string) User {
	var createdAt = time.Now()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return User{
		internal.GetUUID(), createdAt, username, string(passwordHash), displayName}
}
