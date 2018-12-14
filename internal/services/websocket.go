package services

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"speakerbob/internal"
)

const MessageChannel = "messages"

var clients = make(map[*websocket.Conn]bool) // connected clients
var upgrader = websocket.Upgrader{}

// WSConnect: View for upgrading ws requests to ws connections.
func WSConnect(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		return
	}

	clients[ws] = true

	defer ws.Close()
	delete(clients, ws)
}

// SendMessage: Send message to all websocket.
func SendMessage(message Message) {
	internal.GetRedisClient().Publish(MessageChannel, message)
}

// WSMessageConsumer: Go channel for consuming incoming messages from redis and sending them to any active websocket.
func WSMessageConsumer() {
	pubsub := internal.GetRedisClient().Subscribe(MessageChannel)

	_, err := pubsub.Receive()
	if err != nil {
		panic(err)
	}

	// Go channel which receives messages.
	ch := pubsub.Channel()

	for {
		for msg := range ch {
			fmt.Println(msg.Channel, msg.Payload)
			for client := range clients {
				err := client.WriteJSON(msg.Payload)
				if err != nil {
					log.Printf("error: %v", err)
					_ = client.Close()
					delete(clients, client)
				}
			}
		}
	}
}
