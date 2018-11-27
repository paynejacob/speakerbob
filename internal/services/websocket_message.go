package services

type MessageType int

const (
	PLAY_SOUND  MessageType = 0 + iota
	USER_CHANGE
)

// Base class for websocket message.
type Message struct {
	MessageType MessageType
}



type PlaySoundMessage struct  {
	*Message
	Sound string `json:"sound"`
	NSFW bool `json:"nsfw"`
}

func NewPlaySoundMessage(sound string, NSFW bool) *PlaySoundMessage {
	return &PlaySoundMessage{&Message{PLAY_SOUND}, sound, NSFW}
}
