package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/markbates/pkger"
)

type View struct {
	template *template.Template
	layout   string
}

func NewView(layout string, views ...string) (*View, error) {
	lp := filepath.Join("/views/layout", layout+".html")

	paths := []string{lp}
	for _, v := range views {
		paths = append(paths, filepath.Join("/views", v+".html"))
	}

	t := template.New(layout).Funcs(template.FuncMap{
		"date": date,
	})

	for _, p := range paths {
		f, err := pkger.Open(p)
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		if _, err := t.Parse(string(b)); err != nil {
			return nil, err
		}
	}

	return &View{
		template: t,
		layout:   layout,
	}, nil
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.template.ExecuteTemplate(w, v.layout, data)
}

func Must(v *View, err error) *View {
	if err != nil {
		panic(err)
	}
	return v
}

func date(d time.Time) string {
	months := []string{
		"Janvier", "Février", "Mars", "Avril", "Mai", "Juin",
		"Juillet", "Août", "Septembre", "Octobre", "Novembre", "Décembre",
	}
	return fmt.Sprintf("%d %s %d", d.Day(), months[d.Month()-1], d.Year())
}
