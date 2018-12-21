package websocket

import "speakerbob/internal/sound"

type MessageType int

const (
	PLAY_SOUND MessageType = 0 + iota
	PLAY_MACRO
	USER_CHANGE
)

type Message struct {
	MessageType MessageType `json:"type"`
	Channels    ChannelSet  `json:"channels"`
}

func NewPlaySoundMessage(channels ChannelSet, targetSound sound.Sound) *PlaySoundMessage {
	return &PlaySoundMessage{Message{PLAY_SOUND, channels}, targetSound.Id, targetSound.NSFW}
}

type PlaySoundMessage struct {
	Message
	Sound string `json:"sound"`
	NSFW  bool   `json:"nsfw"`
}

func NewPlayMacroMessage(channels ChannelSet, targetMacro sound.Macro) *PlayMacroMessage {
	return &PlayMacroMessage{Message{PLAY_MACRO, channels}, targetMacro.Id, targetMacro.NSFW}
}

type PlayMacroMessage struct {
	Message
	Macro string `json:"sound"`
	NSFW  bool   `json:"nsfw"`
}
