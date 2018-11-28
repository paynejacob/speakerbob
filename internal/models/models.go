package models

import (
	"time"
)

type Sound struct {
	Id        string     `gorm:"primary_key;unique;index" json:"id"`
	CreatedAt *time.Time `json:"create_at"`

	Name      string `gorm:"unique;index" json:"name"`
	Duration  int    `gorm:"default:0" json:"duration"`
	NSFW      bool   `gorm:"default:false" json:"nsfw"`
	Visible   bool   `gorm:"default:true" json:"visible"`
	PlayCount int    `gorm:"default:0" json:"play_count"`
}

type Macro struct {
	Id        string `gorm:"primary_key;unique;index"`
	CreatedAt *time.Time

	Name      string `gorm:"unique;index"`
	PlayCount int    `gorm:"default:0"`
}

type PositionalSound struct {
	Id uint `gorm:"primary_key;unique;index;AUTO_INCREMENT"`

	Position uint

	Sound Sound `gorm:"foreignkey:SoundRefer"`
	Macro Macro `gorm:"foreignkey:MacroRefer"`

	SoundRefer string
	MacroRefer string
}
