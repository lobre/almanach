package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/markbates/pkger"
	"github.com/pkg/browser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	pkger.Include("/views")

	var (
		dbHost = flag.String("db-host", "localhost", "database host")
		dbPort = flag.Int("db-port", 5432, "database port")
		dbUser = flag.String("db-user", "postgres", "database user")
		dbPass = flag.String("db-pass", "postgres", "database password")
		dbName = flag.String("db-name", "postgres", "database name")

		addr        = flag.String("addr", ":8080", "http server address")
		openBrowser = flag.Bool("open-browser", true, "open the browser automatically")
	)
	flag.Parse()

	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	db, err := Open(connStr)
	if err != nil {
		return fmt.Errorf("cannot connect to the db: %w", err)
	}
	defer db.Close()

	logger := log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime)

	app := NewApp(db, logger)
	server := &http.Server{Addr: *addr, Handler: app}

	if *openBrowser {
		host, port, err := net.SplitHostPort(*addr)
		if err != nil {
			return err
		}

		if host == "" {
			host = "localhost"
		}

		url := fmt.Sprintf("http://%s:%s", host, port)

		time.AfterFunc(100*time.Millisecond, func() {
			logger.Print("open app in browser")
			if err := browser.OpenURL(url); err != nil {
				logger.Print("cannot open browser")
			}
		})
	}

	logger.Printf("starting server on: %s", *addr)
	return server.ListenAndServe()
}
