package main

import (
	"html/template"
	"net/http"
)

type View struct {
	template *template.Template
	layout   string
}

func NewView(layout string, files ...string) *View {
	t := template.Must(template.ParseGlob("views/layout/*.html"))
	t = template.Must(t.ParseFiles(files...))

	return &View{
		template: t,
		layout:   layout,
	}
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.template.ExecuteTemplate(w, v.layout, data)
}
