package main

import (
	"log"
	"net/http"
	"os"
)

// App provides application-level context.
// It wraps the routing of the application and
// defines handler to react to specific behaviors.
// To do so, it has a dependency with the DB.
//
// An app is made to execute within an http.Server.
type App struct {
	router *http.ServeMux
	db     *DB
	Logger *log.Logger
}

func NewApp(db *DB) *App {
	app := &App{
		router: http.NewServeMux(),
		db:     db,
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	app.setupRoutes()
	return app
}

func (app *App) setupRoutes() {
	app.router.HandleFunc("/", app.handleIndex())
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.router.ServeHTTP(w, r)
}
