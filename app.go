package main

import (
	"log"
	"net/http"
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
	logger *log.Logger

	indexView *View
	newView   *View
}

func NewApp(db *DB, logger *log.Logger) *App {
	app := &App{
		router: http.NewServeMux(),
		db:     db,
		logger: logger,
	}
	app.setupRoutes()
	app.setupViews()
	return app
}

func (app *App) setupRoutes() {
	app.router.HandleFunc("/", app.withLogs(app.handleIndex()))
	app.router.HandleFunc("/new", app.withLogs(app.handleNew()))
	app.router.HandleFunc("/assets/", app.withLogs(app.handleAssets()))
}

func (app *App) setupViews() {
	app.indexView = Must(NewView("base", "index"))
	app.newView = Must(NewView("base", "new"))
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.router.ServeHTTP(w, r)
}
