package main

//go:generate pkger

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/markbates/pkger"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/browser"
)

func main() {
	logger := log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime)
	if err := run(logger); err != nil {
		logger.Printf("%s\n", err)
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {
	pkger.Include("/views")

	var (
		dbPath      = flag.String("db", "almanach.db", "sqlite database")
		addr        = flag.String("addr", ":8080", "http server address")
		skipBrowser = flag.Bool("skip-browser", false, "don't open the browser automatically")
	)
	flag.Parse()

	db, isNew, err := Open(*dbPath)
	if err != nil {
		return fmt.Errorf("cannot connect to the db: %w", err)
	}
	defer db.Close()

	if isNew {
		logger.Print("new database created, creating schema")
		if err := db.CreateSchema(); err != nil {
			return fmt.Errorf("cannot create db schema: %w", err)
		}
	}

	app := NewApp(db, logger)
	server := &http.Server{Addr: *addr, Handler: app}

	if !(*skipBrowser) {
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
