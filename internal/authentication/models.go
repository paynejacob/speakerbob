package authentication

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
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
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return User{strings.Replace(uuid.New().String(), "-", "", 4), time.Now(), username, string(passwordHash), displayName}
}
