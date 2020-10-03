package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dbHost = flag.String("db-host", "localhost", "database host")
		dbPort = flag.Int("db-port", 5432, "database port")
		dbUser = flag.String("db-user", "postgres", "database user")
		dbPass = flag.String("db-pass", "postgres", "database password")
		dbName = flag.String("db-name", "postgres", "database name")

		sqlImport  = flag.String("sql-import", "", "execute sql file and exit")
		listenAddr = flag.String("listen-addr", ":8080", "http server address")
	)
	flag.Parse()

	logger = log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime)

	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	repo := repo{}

	var err error
	repo.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("cannot connect to the db: %w", err)
	}
	defer repo.db.Close()

	if *sqlImport != "" {
		if err := repo.execFile(*sqlImport); err != nil {
			return fmt.Errorf("can't import sql: %w", err)
		}

		log.Printf("successfully imported %s", *sqlImport)
		os.Exit(0)
	}

	server := &http.Server{
		Addr:         *listenAddr,
		Handler:      withLogger(logger, NewApp(nil)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Printf("server listening on %s", *listenAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("error when starting server: %w", err)

	case <-shutdown:
		logger.Println("start server shutdown")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			<-shutdown
			cancel()
		}()

		if err := server.Shutdown(ctx); err != nil {
			logger.Print("graceful shutdown was interrupted")
			if err = server.Close(); err != nil {
				return fmt.Errorf("error when stopping server: %w", err)
			}
		}
	}

	return nil
}

func withLogger(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("%s: %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
