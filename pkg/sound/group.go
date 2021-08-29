package sound

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

//go:generate go run github.com/paynejacob/hotcereal providergen github.com/paynejacob/speakerbob/pkg/sound.Group
type Group struct {
	Id        string    `json:"id,omitempty" hotcereal:"key"`
	CreatedAt time.Time `json:"created_at,omitempty"`

	Name     string        `json:"name,omitempty" hotcereal:"searchable"`
	Duration time.Duration `json:"duration,omitempty"`
	SoundIds []string      `json:"sounds,omitempty"`
}

func NewGroup() Group {
	return Group{
		Id:        strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt: time.Now(),
	}
}
