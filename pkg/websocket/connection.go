package websocket

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// max messages to buffer per connection
	sendChannelSize = 256
)

type Conn struct {
	ws        *websocket.Conn
	ExpiresAt time.Time

	service *Service
	send    chan interface{}
}

func NewConn(ws *websocket.Conn, service *Service) *Conn {
	return &Conn{ws: ws, service: service, send: make(chan interface{}, sendChannelSize)}
}

func (c *Conn) SendMessage(msg interface{}) {
	c.send <- msg
}

func (c *Conn) readPump() {
	defer func() {
		c.service.unRegisterConnection(c)
		_ = c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	_ = c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { return c.ws.SetReadDeadline(time.Now().Add(pongWait)) })

	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.ws.Close()
	}()

	for {
		select {
		case message := <-c.send:
			_ = c.ws.WriteJSON(message)
		case <-ticker.C:
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
