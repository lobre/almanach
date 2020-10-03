package main

import "net/http"

func (app *App) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}
}
