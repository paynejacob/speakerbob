package api

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"hash/fnv"
	"strings"
)

type Connection struct {
	id       string
	ws       *websocket.Conn
	user     User
	nsfw     bool
	channels ChannelSet
}

func (c *Connection) Hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(c.id))

	return h.Sum64()
}

type Channel struct {
	Value string
}

func (c *Channel) Hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(c.Value))

	return h.Sum64()
}

type WebsocketBackend interface {
	Channel() chan IMessage
	SendMessage(message IMessage)
	Connections(channels ChannelSet) []*Connection
	LocalConnections(channels ChannelSet) []*Connection
	RegisterConnection(ws *websocket.Conn, user User, channels ChannelSet, nsfw bool) *Connection
	CloseConnection(connection *Connection)
}

type WebsocketMemoryBackend struct {
	channel     chan IMessage
	connections ConnectionSet
	channelMap  map[string]ConnectionSet
}

func NewWebsocketMemoryBackend() WebsocketMemoryBackend {
	return WebsocketMemoryBackend{make(chan IMessage), ConnectionSet{}, make(map[string]ConnectionSet)}
}

func (b WebsocketMemoryBackend) Channel() chan IMessage {
	return b.channel
}

func (b WebsocketMemoryBackend) SendMessage(message IMessage) {
	b.channel <- message
}

func (b WebsocketMemoryBackend) Connections(channels ChannelSet) []*Connection {
	var connections ConnectionSet

	for _, channel := range channels {
		if conns, ok := b.channelMap[channel.Value]; ok {
			connections = connections.Union(conns)
		}
	}

	return connections.Values()
}

func (b WebsocketMemoryBackend) LocalConnections(channels ChannelSet) []*Connection {
	return b.Connections(channels)
}

func (b WebsocketMemoryBackend) RegisterConnection(ws *websocket.Conn, user User, channels ChannelSet, nsfw bool) *Connection {
	connection := &Connection{strings.Replace(uuid.New().String(), "-", "", 4), ws, user, nsfw, channels}

	for _, channel := range channels.Values() {
		if _, ok := b.channelMap[channel.Value]; !ok {
			newSet := ConnectionSet{}
			b.channelMap[channel.Value] = newSet
		}
		b.channelMap[channel.Value].Add(connection)
	}
	b.connections.Add(connection)

	return connection
}

func (b WebsocketMemoryBackend) CloseConnection(connection *Connection) {
	for _, channel := range connection.channels.Values() {
		b.channelMap[channel.Value].Remove(connection)
		if len(b.channelMap[channel.Value]) == 0 {
			delete(b.channelMap, channel.Value)
		}
	}

	b.connections.Remove(connection)
}
