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

		addr   = flag.String("addr", ":8080", "http server address")
		domain = flag.String("domain", "", "activate self signed or letsencrypt certificate on domain")
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
	if *domain != "" {
		srv.ConfigureTLS(*domain)
	}
	return srv.ServeUntilSignal()
}
