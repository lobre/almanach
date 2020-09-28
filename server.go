package main

import (
	"fmt"
	"net/http"
)

type server struct {
	repo   *repo
	router *http.ServeMux
}

func newServer(r *repo) *server {
	// TODO: wrap the router handler
	// with a middleware that would log

	return &server{
		repo:   r,
		router: http.NewServeMux(),
	}
}

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleIndex())
}

func (s *server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	}
}
