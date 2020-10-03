package main

import (
	"log"
	"net/http"
)

type App struct {
	router *http.ServeMux
	db     *DB
	logger *log.Logger
}

func NewApp(db *DB, logger *log.Logger) *App {
	app := &App{
		router: http.NewServeMux(),
		db:     db,
		logger: logger,
	}
	app.setupRoutes()
	return app
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.router.ServeHTTP(w, r)
}
