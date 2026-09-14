package websocket

type MessageType string

const (
	PlayMessageType            = "play"
	UpdateSoundMessageType     = "update_sound"
	DeleteSoundMessageType     = "delete_sound"
	CreateGroupMessageType     = "create_group"
	UpdateGroupMessageType     = "update_group"
	DeleteGroupMessageType     = "delete_group"
	ConnectionCountMessageType = "connection_count"
	PresenceMessageType        = "presence"
)

type ConnectionCountMessage struct {
	Type  MessageType `json:"type"`
	Count int         `json:"count"`
}

type PresenceMessage struct {
	Type  MessageType `json:"type"`
	Users []string    `json:"users"`
}
