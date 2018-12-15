package authentication

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
		uuid.New().String(), createdAt, username, string(passwordHash), displayName}
}
