package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbHost := flag.String("db-host", "localhost", "database host")
	dbPort := flag.Int("db-port", 5432, "database port")
	dbUser := flag.String("db-user", "postgres", "database user")
	dbPass := flag.String("db-pass", "postgres", "database password")
	dbName := flag.String("db-name", "postgres", "database name")

	sqlImport := flag.String("sql-import", "", "execute sql file and exit")

	httpAddr := flag.String("http-addr", ":8080", "http server address")

	flag.Parse()

	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	repo := repo{}

	var err error
	repo.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("cannot connect to the db: %v", err)
	}
	defer repo.db.Close()

	if *sqlImport != "" {
		if err := repo.execFile(*sqlImport); err != nil {
			log.Fatalf("can't import sql: %v", err)
		}

		log.Printf("successfully imported %s", *sqlImport)
		os.Exit(0)
	}

	srv := newServer(&repo)
	srv.routes()

	if err := http.ListenAndServe(*httpAddr, srv.router); err != nil {
		log.Fatalf("cannot start http server: %v", err)
	}
}
