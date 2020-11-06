package main

import (
	"net/http"

	"github.com/markbates/pkger"
)

func (app *App) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := app.db.getEvents()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		app.indexView.Render(w, map[string]interface{}{
			"Events": events,
		})
	}
}

func (app *App) handleAssets() http.HandlerFunc {
	fs := http.FileServer(pkger.Dir("/assets"))
	return func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", fs).ServeHTTP(w, r)
	}
}
