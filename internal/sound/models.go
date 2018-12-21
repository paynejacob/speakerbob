package sound

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type Sound struct {
	Id        string    `gorm:"primary_key;unique;index" json:"id"`
	CreatedAt time.Time `json:"create_at"`

	Name      string `gorm:"unique;index" json:"name"`
	Duration  int    `gorm:"default:0" json:"duration"`
	NSFW      bool   `gorm:"default:false" json:"nsfw"`
	Visible   bool   `gorm:"default:true" json:"visible"`
	PlayCount int    `gorm:"default:0" json:"play_count"`
}

func NewSound(name string, nsfw bool, visible bool) Sound {
	return Sound{strings.Replace(uuid.New().String(), "-", "", 4), time.Now(), name, 0, nsfw, visible, 0}
}

type Macro struct {
	Id        string     `gorm:"primary_key;unique;index" json:"id"`
	CreatedAt *time.Time `json:"created_at"`

	Name      string `gorm:"unique;index" json:"name"`
	PlayCount int    `gorm:"default:0" json:"play_count"`

	NSFW bool `gorm:"-"`
}

type PositionalSound struct {
	Id uint `gorm:"primary_key;unique;index;AUTO_INCREMENT"`

	Position uint

	Sound Sound `gorm:"foreignkey:SoundRefer"`
	Macro Macro `gorm:"foreignkey:MacroRefer"`

	SoundRefer string
	MacroRefer string
}
