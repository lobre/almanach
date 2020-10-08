package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// Server is a wrapper around http.Server
// that bring graceful shutdown and easy
// tls configuration.
type Server struct {
	*http.Server
	httpsEnforcer *http.Server

	logger *log.Logger
}

// NewServer instanciates a Server.
func NewServer(addr string, h http.Handler, logger *log.Logger) *Server {
	srv := Server{logger: logger}

	srv.Server = &http.Server{
		Addr:         addr,
		Handler:      srv.withAccessLogs(h),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
		ErrorLog:     logger,
	}

	return &srv
}

// ConfigureTLS will make the server run with a tls encryption.
//
// If acme is true, a certificate will be generated with letsencrypt.
//
// Otherwise, the certificate and key should be present under
// "certs/<domain>.crt" and "certs/<domain>.key".
func (srv *Server) ConfigureTLS(domain string, acme bool, httpsOnly bool) {
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("certs"),
	}

	srv.TLSConfig = certManager.TLSConfig()
	srv.TLSConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if acme {
			return certManager.GetCertificate(hello)
		}

		dirCache, ok := certManager.Cache.(autocert.DirCache)
		if !ok {
			dirCache = "certs"
		}

		keyFile := filepath.Join(string(dirCache), hello.ServerName+".key")
		crtFile := filepath.Join(string(dirCache), hello.ServerName+".crt")
		cert, err := tls.LoadX509KeyPair(crtFile, keyFile)

		return &cert, err
	}

	if httpsOnly {
		srv.httpsEnforcer = &http.Server{
			Addr:         ":80",
			Handler:      srv.withAccessLogs(certManager.HTTPHandler(nil)),
			ReadTimeout:  srv.ReadTimeout,
			WriteTimeout: srv.WriteTimeout,
			IdleTimeout:  srv.IdleTimeout,
			ErrorLog:     srv.logger,
		}
	}
}

// ServeUntilSignal starts the embedded http.Server and waits for
// an exit signal to gracefully shutdown.
//
// If tls has been configured with Server.ConfigureTLS using the
// httpsOnly option, it also spins up an additional server on port 80
// to redirect requests from http to https.
func (srv *Server) ServeUntilSignal() error {
	serverErrors := make(chan error, 1)

	go func() {
		if srv.TLSConfig != nil {
			srv.logger.Printf("starting server with tls on %s", srv.Addr)
			serverErrors <- srv.ListenAndServeTLS("", "")
		} else {
			srv.logger.Printf("starting server on %s", srv.Addr)
			serverErrors <- srv.ListenAndServe()
		}
	}()

	if srv.httpsEnforcer != nil {
		go func() {
			srv.logger.Printf("starting https enforcer on %s", srv.httpsEnforcer.Addr)
			serverErrors <- srv.httpsEnforcer.ListenAndServe()
		}()
	}

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

		if err := srv.Shutdown(ctx); err != nil {
			srv.logger.Print("graceful shutdown of server was interrupted")
			if err = srv.Close(); err != nil {
				return fmt.Errorf("error when stopping server: %w", err)
			}
		}

		if srv.httpsEnforcer != nil {
			if err := srv.httpsEnforcer.Shutdown(ctx); err != nil {
				srv.logger.Print("graceful shutdown of https enforcer was interrupted")
				if err = srv.httpsEnforcer.Close(); err != nil {
					return fmt.Errorf("error when stopping https enforcer: %w", err)
				}
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
