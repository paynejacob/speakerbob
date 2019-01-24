package api

import (
	"speakerbob/internal"
	"time"
)

type Sound struct {
	Id        string    `gorm:"primary_key;unique;index" json:"id"`
	CreatedAt time.Time `json:"create_at"`

	Name      string `gorm:"unique;index" json:"name"`
	Duration  int    `gorm:"default:0" json:"duration"`
	NSFW      bool   `gorm:"default:false" json:"nsfw"`
	Visible   bool   `gorm:"default:false" json:"visible"`
	PlayCount int    `gorm:"default:0" json:"play_count"`
}

func NewSound(name string, nsfw bool, visible bool) Sound {
	return Sound{internal.NewId(), time.Now(), name, 0, nsfw, visible, 0}
}

func (Sound) Type() string {
	return "sound"
}

func (r Sound) Key() string {
	return r.Id
}

func (r Sound) IndexValue() string {
	return r.Name
}

func (r Sound) Object() interface{} {
	return r
}

type Macro struct {
	Id        string     `gorm:"primary_key;unique;index" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Name      string `gorm:"unique;index" json:"name"`
	PlayCount int    `gorm:"default:0" json:"play_count"`

	NSFW bool `gorm:"default:false" json:"nsfw"`
}

func NewMacro(name string) *Macro {
	return &Macro{Id: internal.NewId(), CreatedAt: time.Now(), Name: name}
}

func (Macro) Type() string {
	return "sound"
}

func (r Macro) Key() string {
	return r.Id
}

func (r Macro) IndexValue() string {
	return r.Name
}

func (r Macro) Object() interface{} {
	return r
}

type PositionalSound struct {
	Id uint `gorm:"primary_key;unique;index;AUTO_INCREMENT"`

	Position int

	Sound Sound `gorm:"foreignkey:SoundRefer"`
	Macro Macro `gorm:"foreignkey:MacroRefer"`

	SoundRefer string
	MacroRefer string
}

func NewPositionalSound(position int, sound Sound, macro Macro) *PositionalSound {
	return &PositionalSound{Position: position, Sound: sound, Macro: macro, SoundRefer: sound.Id, MacroRefer: macro.Id}
}