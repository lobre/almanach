package main

func (app *App) setupRoutes() {
	app.router.HandleFunc("/", app.handleIndex())
}
