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

	logger *log.Logger
}

func NewServer(addr string, h http.Handler, logger *log.Logger) *Server {
	var (
		readTimeout  = 5 * time.Second
		writeTimeout = 10 * time.Second
		idleTimeout  = 15 * time.Second
	)

	srv := Server{logger: logger}

	srv.mainServer = &http.Server{
		Addr:         addr,
		Handler:      srv.withAccessLogs(h),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     logger,
	}

	srv.redirectServer = &http.Server{
		Addr:         ":80",
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     logger,
	}

	return &srv
}

// ConfigureTLS will make the server run with a TLS encryption.
//
// If a local domain is used (localhost, *.dev, *.local), a self
// signed certificate will be loaded at the path "certs/<domain>.crt"
// alongside with a key at "certs/<domain>.key".
//
// Otherwise, letsencrypt will be used to generate a proper certificate.
func (srv *Server) ConfigureTLS(domain string) {
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("certs"),
	}

	srv.mainServer.TLSConfig = certManager.TLSConfig()
	srv.mainServer.TLSConfig.GetCertificate = getSelfSignedOrLetsEncryptCert(certManager)

	srv.redirectServer.Handler = srv.withAccessLogs(certManager.HTTPHandler(nil))
}

// ServeUntilSignal starts the embedded http.Server and
// waits for an exit signal to gracefully shutdown.
//
// If tls has been configured with Server.ConfigureTLS and the listening port
// is 443, it also spins up an additional server to redirect requests from 80 to 443.
func (srv *Server) ServeUntilSignal() error {
	_, port, err := net.SplitHostPort(srv.mainServer.Addr)
	if err != nil {
		return err
	}

	serverErrors := make(chan error, 1)
	redirectErrors := make(chan error, 1)

	go func() {
		if srv.mainServer.TLSConfig != nil {
			srv.logger.Printf("starting server with tls on %s", srv.mainServer.Addr)
			serverErrors <- srv.mainServer.ListenAndServeTLS("", "")
		} else {
			srv.logger.Printf("starting server on %s", srv.mainServer.Addr)
			serverErrors <- srv.mainServer.ListenAndServe()
		}
	}()

	if srv.mainServer.TLSConfig != nil && port == "443" {
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

func (srv *Server) withAccessLogs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.logger.Printf("%s: %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
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

func isLocal(host string) bool {
	domains := []string{".dev", ".local"}
	for _, domain := range domains {
		if strings.HasSuffix(host, domain) {
			return true
		}
	}
	return host == "localhost"
}
