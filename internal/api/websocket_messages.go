package api

import "encoding/json"

type MessageType int

const (
	PLAY_SOUND MessageType = 0 + iota
	PLAY_MACRO
	USER_CHANGE
)

type IMessage interface {
	Bytes() ([]byte, error)
	Channels() ChannelSet
}

type Message struct {
	MessageType MessageType `json:"type"`
	channels    ChannelSet  `json:"channels"`
}

func (m Message) Bytes() ([]byte, error) {
	if b, err := json.Marshal(m); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

func (m Message) Channels() ChannelSet {
	return m.channels
}

type PlaySoundMessage struct {
	Message
	Sound string `json:"sound"`
	NSFW  bool   `json:"nsfw"`
}

func NewPlaySoundMessage(channels ChannelSet, soundId string, nsfw bool) PlaySoundMessage {
	return PlaySoundMessage{Message{PLAY_SOUND, channels}, soundId, nsfw}
}

type PlayMacroMessage struct {
	Message
	Macro string `json:"sound"`
	NSFW  bool   `json:"nsfw"`
}

func NewPlayMacroMessage(channels ChannelSet, macroId string, nsfw bool) PlayMacroMessage {
	return PlayMacroMessage{Message{PLAY_MACRO, channels}, macroId, nsfw}
}
