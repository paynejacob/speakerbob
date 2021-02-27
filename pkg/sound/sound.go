package sound

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type Sound struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	NSFW     bool          `json:"nsfw"`
}

func NewSound() Sound {
	return Sound{
		Id:        strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt: time.Now(),
	}
}
