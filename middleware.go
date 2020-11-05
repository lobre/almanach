package main

import "net/http"

func (app *App) withLogs(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.logger.Printf("%s: %s %s", r.RemoteAddr, r.Method, r.URL)
		next(w, r)
	}
}
