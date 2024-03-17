package pkg

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type client struct {
	isClosing bool
	mu        sync.Mutex
}

type NotifyClientData struct {
	Conn    *websocket.Conn
	Message []byte
}

var clients = make(map[*websocket.Conn]*client)
var Register = make(chan *websocket.Conn)
var Unregister = make(chan *websocket.Conn)
var Notify = make(chan NotifyClientData)
var Broadcast = make(chan []byte)

type EventBody struct {
	Event    string `json:"event"`
	Contents string `json:"contents,omitempty"`
}

func (app *App) RunHub() {
	for {
		select {
		case connection := <-Register:
			clients[connection] = &client{}

		case message := <-Notify:
			c, ok := clients[message.Conn]
			if !ok {
				log.Println("client not on map")
				Unregister <- message.Conn
				return
			}

			go WriteToConnection(message.Conn, c, message.Message)

		case message := <-Broadcast:
			// Send the message to all clients
			for connection, c := range clients {
				go WriteToConnection(connection, c, message)
			}

		case connection := <-Unregister:
			// Remove the client from the hub
			delete(clients, connection)
		}
	}
}

func WriteToConnection(conn *websocket.Conn, c *client, msg []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosing {
		log.Println("is closing")
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Println("write error:", err)

		c.isClosing = true
		Unregister <- conn

		log.Println("write error:", err)
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
	}
}
