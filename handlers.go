package main

import (
	"fmt"
	"net/http"
	"time"

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

func (app *App) handleNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.newView.Render(w, nil)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			name := r.FormValue("name")
			if name == "" {
				app.logger.Println("Event name not defined")
				return
			}

			loc, err := time.LoadLocation("Europe/Paris")
			if err != nil {
				app.logger.Println("Can't find timezone")
				return
			}

			datetime := fmt.Sprintf("%sT%s", r.FormValue("date"), r.FormValue("time"))
			date, err := time.ParseInLocation("2006-01-02T15:04", datetime, loc)
			if err != nil {
				app.logger.Println("Event date not properly formated or not defined")
				return
			}

			e := Event{
				Name:    name,
				Date:    date,
				Comment: r.FormValue("comment"),
			}

			if _, err := app.db.insertEvent(e); err != nil {
				app.logger.Printf("Can't create event: %s", err)
				return
			}

			http.Redirect(w, r, "/", http.StatusFound)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	}
}

func (app *App) handleAssets() http.HandlerFunc {
	fs := http.FileServer(pkger.Dir("/assets"))
	return func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", fs).ServeHTTP(w, r)
	}
}
