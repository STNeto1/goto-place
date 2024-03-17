package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/stneto1/goto-place/pkg"
	"github.com/stneto1/goto-place/ui"
)

var hexColorPattern = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

func main() {
	// FLY.IO injected value
	primaryRegion := os.Getenv("PRIMARY_REGION")

	dist, err := fs.Sub(ui.Dist, "dist")
	if err != nil {
		log.Fatalln("failed to get fs", err)
	}
	appState := pkg.NewApp(primaryRegion != "")
	defer appState.CloseConnection()

	app := fiber.New(fiber.Config{})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("appState", appState)
		return c.Next()
	})

	app.Get("/ws", websocket.New(WsHandler, websocket.Config{}))
	app.Use("/", filesystem.New(filesystem.Config{
		Root: http.FS(dist),
	}))

	// Access file "image.png" under `static/` directory via URL: `http://<server>/static/image.png`.
	// Without `PathPrefix`, you have to access it via URL:
	// `http://<server>/static/static/image.png`.
	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(ui.Dist),
		PathPrefix: "static",
		Browse:     true,
	}))

	go appState.RunHub()

	app.Listen(":3000")
}

func WsHandler(ws *websocket.Conn) {
	defer func() {
		pkg.Unregister <- ws
		ws.Close()
	}()
	pkg.Register <- ws

	appState, ok := ws.Locals("appState").(*pkg.App)
	if !ok {
		return
	}

	rows, err := appState.GetPoints()
	if err != nil {
		log.Println("failed to get existing points", err)
		return
	}

	jsonData, err := json.Marshal(rows)
	if err != nil {
		log.Println("failed to serialize", err)
		return
	}

	pkg.Notify <- pkg.NotifyClientData{Conn: ws, Message: jsonData}

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

				if err := appState.UpdatePoint(x, y, hexColor); err != nil {
					log.Println("failed to update", err)
					continue
				}

				pkg.Broadcast <- []byte(fmt.Sprintf("\"%d;%d;%s\"", x, y, hexColor))

				continue
			}

			log.Println("payload for weird event", payload.Event, "with contents:", payload.Contents)
			continue
		}

		log.Println("got message of the type", msgType, "with the content:", string(msg))

	}
}
