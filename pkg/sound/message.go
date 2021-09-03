package sound

import (
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"time"
)

type PlayMessage struct {
	Type      websocket.MessageType `json:"type"`
	Sound     Sound                 `json:"sound"`
	Scheduled time.Time             `json:"scheduled"`
}

type SoundMessage struct {
	Type  websocket.MessageType `json:"type"`
	Sound *Sound                `json:"sound"`
}

type GroupMessage struct {
	Type  websocket.MessageType `json:"type"`
	Group *Group                `json:"group"`
}
