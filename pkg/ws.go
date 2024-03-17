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

var Clients = make(map[*websocket.Conn]*client)
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
			Clients[connection] = &client{}

		case message := <-Notify:
			c, ok := Clients[message.Conn]
			if !ok {
				log.Println("client not on map")
				Unregister <- message.Conn
				return
			}
			go func(client *client, msg *NotifyClientData) {
				client.mu.Lock()
				defer client.mu.Unlock()

				if client.isClosing {
					return
				}

				if err := WriteToConnection(msg.Conn, msg.Message); err != nil {
					log.Println("write error:", err)

					client.isClosing = true
					Unregister <- msg.Conn
				}
			}(c, &message)

		case message := <-Broadcast:
			// Send the message to all clients
			for connection, c := range Clients {
				go func(connection *websocket.Conn, c *client) { // send to each client in parallel so we don't block on a slow client
					c.mu.Lock()
					defer c.mu.Unlock()

					if c.isClosing {
						log.Println("is closing")
						return
					}

					if err := WriteToConnection(connection, message); err != nil {
						log.Println("write error:", err)

						c.isClosing = true
						Unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-Unregister:
			// Remove the client from the hub
			delete(Clients, connection)
		}
	}
}

func WriteToConnection(conn *websocket.Conn, msg []byte) error {
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Println("write error:", err)

		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
	}

	return nil
}
