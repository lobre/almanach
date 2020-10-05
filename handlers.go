package main

import (
	"net/http"
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
