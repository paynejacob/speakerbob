package api

import "encoding/json"

type MessageType int

const (
	PLAY_SOUND MessageType = 0 + iota
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

func NewPlaySoundMessage(channels ChannelSet, sound Sound) PlaySoundMessage {
	return PlaySoundMessage{Message{PLAY_SOUND, channels}, sound.Id, sound.NSFW}
}
