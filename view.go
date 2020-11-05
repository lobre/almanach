package main

import (
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/markbates/pkger"
)

type View struct {
	template *template.Template
	layout   string
}

func NewView(layout string, filenames ...string) *View {
	t := template.New(layout)
	s := loadTemplate("/views/layout/" + layout + ".html")
	template.Must(t.Parse(s))

	for _, filename := range filenames {
		s = loadTemplate(filename)
		template.Must(t.Parse(s))
	}

	return &View{
		template: t,
		layout:   layout,
	}
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.template.ExecuteTemplate(w, v.layout, data)
}

func loadTemplate(filename string) string {
	f, err := pkger.Open(filename)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return string(b)
}
