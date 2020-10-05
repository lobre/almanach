package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	*http.Server
	logger *log.Logger
}

func NewServer(listenAddr string, h http.Handler, logger *log.Logger) *Server {
	srv := Server{logger: logger}
	srv.Server = &http.Server{
		Addr:         listenAddr,
		Handler:      srv.withAccessLogs(h),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	return &srv
}

// ServeUntilSignal starts the embedded http.Server and
// waits for an exit signal to gracefully shutdown.
func (srv *Server) ServeUntilSignal() error {
	serverErrors := make(chan error, 1)
	go func() {
		srv.logger.Printf("server listening on %s", srv.Addr)
		serverErrors <- srv.Server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("error when starting server: %w", err)

	case <-shutdown:
		srv.logger.Println("start server shutdown")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			<-shutdown
			cancel()
		}()

		if err := srv.Server.Shutdown(ctx); err != nil {
			srv.logger.Print("graceful shutdown was interrupted")
			if err = srv.Server.Close(); err != nil {
				return fmt.Errorf("error when stopping server: %w", err)
			}
		}
	}

	return nil
}

func (srv *Server) withAccessLogs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.logger.Printf("%s: %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
