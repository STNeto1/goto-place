package pkg

import (
	"database/sql"
	"log"

	_ "github.com/superfly/litefs-go"
	_ "github.com/tursodatabase/go-libsql"
)

func NewApp(isProduction bool) *App {
	if isProduction {
		conn, err := sql.Open("sqlite3", "/litefs/my.db")
		if err != nil {
			log.Panicln("Failed to open connection", err)
		}
		if err := conn.Ping(); err != nil {
			log.Panicln("Failed to ping", err)
		}

		return &App{conn}
	}

	conn, err := sql.Open("libsql", "file:db.sqlite")
	if err != nil {
		log.Panicln("Failed to open connection", err)
	}
	if err := conn.Ping(); err != nil {
		log.Panicln("Failed to ping", err)
	}

	return &App{conn}
}

type App struct {
	conn *sql.DB
}

func (app *App) CloseConnection() {
	app.conn.Close()
}

type Point struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"c"`
}

func (app *App) GetPoints() ([]Point, error) {
	rows, err := app.conn.Query("SELECT x, y, color FROM points")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []Point
	for rows.Next() {
		var point Point
		if err := rows.Scan(&point.X, &point.Y, &point.Color); err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	return points, nil
}

var query = `INSERT INTO points (x, y, color)
VALUES (?, ?, ?)
ON CONFLICT DO UPDATE SET color = ?;`

func (app *App) UpdatePoint(x, y int, color string) error {
	_, err := app.conn.Exec(query, x, y, color, color)
	return err
}
