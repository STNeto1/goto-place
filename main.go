package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/stneto1/goto-place/pkg"
)

var hexColorPattern = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

func main() {
	appState := pkg.NewApp()

	app := fiber.New(fiber.Config{})
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	app.Static("/public", "./public")
	app.Get("/ws", websocket.New(WsHandler, websocket.Config{}))

	go appState.RunHub()

	app.Listen(":3000")
}

func WsHandler(ws *websocket.Conn) {
	defer func() {
		pkg.Unregister <- ws
		ws.Close()
	}()

	pkg.Register <- ws

	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}
			break
		}

		if msgType == websocket.CloseMessage {
			break
		}

		if msgType == websocket.TextMessage {
			var payload pkg.EventBody
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Println("error unmarshalling payload:", err)
				continue
			}

			if payload.Event == "click" {
				tokens := strings.Split(payload.Contents, ";")
				if len(tokens) != 3 {
					pkg.Notify <- pkg.NotifyClientData{Conn: ws, Message: []byte("invalid click event")}
					log.Println("invalid click event:", payload.Contents)
					continue
				}

				x, err := strconv.Atoi(tokens[0])
				if err != nil {
					pkg.Notify <- pkg.NotifyClientData{Conn: ws, Message: []byte(fmt.Sprintf("invalid x value: %s", tokens[0]))}
					// pkg.Broadcast <- []byte("invalid x")
					log.Println("invalid x:", tokens[0])
					continue
				}

				y, err := strconv.Atoi(tokens[1])
				if err != nil {
					pkg.Notify <- pkg.NotifyClientData{Conn: ws, Message: []byte(fmt.Sprintf("invalid y value: %s", tokens[1]))}
					// pkg.Broadcast <- []byte("invalid y")
					log.Println("invalid y:", tokens[1])
					continue
				}

				hexColor := tokens[2]
				if !hexColorPattern.MatchString(hexColor) {
					pkg.Notify <- pkg.NotifyClientData{Conn: ws, Message: []byte(fmt.Sprintf("invalid hex color: %s", tokens[2]))}
					// pkg.Broadcast <- []byte("invalid color")
					log.Println("invalid color:", hexColor)
					continue
				}

				pkg.Broadcast <- []byte(fmt.Sprintf("%d;%d;%s", x, y, hexColor))

				continue
			}

			log.Println("payload for weird event", payload.Event, "with contents:", payload.Contents)
		}

		log.Println("got message of the type", msgType, "with the content:", string(msg))

	}
}
