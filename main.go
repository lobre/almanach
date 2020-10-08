package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
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

		addr = flag.String("addr", ":8080", "http server address")

		tls       = flag.Bool("tls", false, "activate tls encryption")
		tlsDomain = flag.String("tls-domain", "localhost", "domain used for tls certificate")
		tlsAcme   = flag.Bool("tls-acme", false, "activate generation of certificate with letsencrypt")
		httpsOnly = flag.Bool("https-only", false, "listen on port 80 to redirect http requests to https")
	)
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime)

	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	db, err := Open(connStr)
	if err != nil {
		return fmt.Errorf("cannot connect to the db: %w", err)
	}
	defer db.Close()

	app := NewApp(db, logger)
	srv := NewServer(*addr, app, logger)
	if *tls {
		srv.ConfigureTLS(*tlsDomain, *tlsAcme, *httpsOnly)
	}
	return srv.ServeUntilSignal()
}
