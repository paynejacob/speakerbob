package websocket

type MessageType string

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}
