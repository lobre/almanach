package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	mainServer     *http.Server
	redirectServer *http.Server

	host       string
	port       string
	disableTLS bool

	logger *log.Logger
}

func NewServer(host string, port string, h http.Handler, disableTLS bool, logger *log.Logger) *Server {
	var (
		readTimeout  = 5 * time.Second
		writeTimeout = 10 * time.Second
		idleTimeout  = 15 * time.Second
	)

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "443"
	}

	srv := Server{
		host:       host,
		port:       port,
		disableTLS: disableTLS,
		logger:     logger,
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Cache:      autocert.DirCache("certs"),
	}

	tlsConfig := certManager.TLSConfig()
	tlsConfig.GetCertificate = getSelfSignedOrLetsEncryptCert(certManager)

	srv.mainServer = &http.Server{
		Addr:         net.JoinHostPort(host, port),
		Handler:      srv.withAccessLogs(h),
		TLSConfig:    tlsConfig,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     logger,
	}

	srv.redirectServer = &http.Server{
		Addr:         net.JoinHostPort(host, "80"),
		Handler:      srv.withAccessLogs(certManager.HTTPHandler(nil)),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     logger,
	}

	return &srv
}

// ServeUntilSignal starts the embedded http.Server and
// waits for an exit signal to gracefully shutdown.
//
// If the server is in TLS mode and the listening port is 443,
// it also spins up an additional server redirecting requests
// from HTTP to HTTPS.
func (srv *Server) ServeUntilSignal() error {
	serverErrors := make(chan error, 1)
	redirectErrors := make(chan error, 1)

	go func() {
		srv.logger.Printf("starting server on %s", srv.mainServer.Addr)
		if srv.disableTLS {
			serverErrors <- srv.mainServer.ListenAndServe()
		} else {
			serverErrors <- srv.mainServer.ListenAndServeTLS("", "")
		}
	}()

	if !srv.disableTLS && srv.port == "443" {
		go func() {
			srv.logger.Printf("starting redirect server on %s", srv.redirectServer.Addr)
			redirectErrors <- srv.redirectServer.ListenAndServe()
		}()
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("error when starting server: %w", err)

	case err := <-redirectErrors:
		return fmt.Errorf("error when starting redirect server: %w", err)

	case <-shutdown:
		srv.logger.Println("start server shutdown")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			<-shutdown
			cancel()
		}()

		if err := srv.mainServer.Shutdown(ctx); err != nil {
			srv.logger.Print("graceful shutdown of server was interrupted")
			if err = srv.mainServer.Close(); err != nil {
				return fmt.Errorf("error when stopping server: %w", err)
			}
		}

		if err := srv.redirectServer.Shutdown(ctx); err != nil {
			srv.logger.Print("graceful shutdown of redirect server was interrupted")
			if err = srv.redirectServer.Close(); err != nil {
				return fmt.Errorf("error when stopping redirect server: %w", err)
			}
		}
	}

	return nil
}

func getSelfSignedOrLetsEncryptCert(certManager *autocert.Manager) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if !isLocal(hello.ServerName) {
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
}

func (srv *Server) withAccessLogs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.logger.Printf("%s: %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func isLocal(host string) bool {
	domains := []string{".dev", ".local"}
	for _, domain := range domains {
		if strings.HasSuffix(host, domain) {
			return true
		}
	}
	return host == "localhost"
}
