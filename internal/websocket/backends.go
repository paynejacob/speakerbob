package websocket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"hash/fnv"
	"speakerbob/internal/authentication"
	"strings"
)

type Connection struct {
	id       string
	ws       *websocket.Conn
	user     authentication.User
	nsfw     bool
	channels ChannelSet
}

func (c *Connection) Hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(c.id))

	return h.Sum64()
}

type Channel struct {
	value string
}

func (c *Channel) Hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(c.value))

	return h.Sum64()
}

type Backend interface {
	Channel() chan Message
	SendMessage(message Message)
	Connections(channels ChannelSet) []*Connection
	LocalConnections(channels ChannelSet) []*Connection
	RegisterConnection(ws *websocket.Conn, user authentication.User, channels ChannelSet, nsfw bool) *Connection
	CloseConnection(connection *Connection)
}

func NewMemoryBackend() MemoryBackend {
	return MemoryBackend{}
}

type MemoryBackend struct {
	channel     chan Message
	connections ConnectionSet
	channelMap  map[string]ConnectionSet
}

func (b MemoryBackend) Channel() chan Message {
	return b.channel
}

func (b MemoryBackend) SendMessage(message Message) {
	b.channel <- message
}

func (b MemoryBackend) Connections(channels ChannelSet) []*Connection {
	var connections ConnectionSet

	for _, channel := range channels {
		if conns, ok := b.channelMap[channel.value]; ok {
			connections = connections.Union(conns)
		}
	}

	return connections.Values()
}

func (b MemoryBackend) LocalConnections(channels ChannelSet) []*Connection {
	return b.Connections(channels)
}

func (b MemoryBackend) RegisterConnection(ws *websocket.Conn, user authentication.User, channels ChannelSet, nsfw bool) *Connection {
	connection := &Connection{strings.Replace(uuid.New().String(), "-", "", 4), ws, user, nsfw, channels}

	for _, channel := range channels.Values() {
		if _, ok := b.channelMap[channel.value]; !ok {
			newSet := ConnectionSet{}
			b.channelMap[channel.value] = newSet
		}
		b.channelMap[channel.value].Add(connection)
	}
	b.connections.Add(connection)

	return connection
}

func (b MemoryBackend) CloseConnection(connection *Connection) {
	for _, channel := range connection.channels.Values() {
		b.channelMap[channel.value].Remove(connection)
		if len(b.channelMap[channel.value]) == 0 {
			delete(b.channelMap, channel.value)
		}
	}

	b.connections.Remove(connection)
}
