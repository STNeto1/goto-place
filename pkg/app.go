package pkg

import "database/sql"

type App struct {
	conn *sql.DB
}

func NewApp() *App {
	return nil
}
