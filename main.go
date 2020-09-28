package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dbHost := flag.String("db-host", "localhost", "database host")
	dbPort := flag.Int("db-port", 5432, "database port")
	dbUser := flag.String("db-user", "postgres", "database user")
	dbPass := flag.String("db-pass", "postgres", "database password")
	dbName := flag.String("db-name", "postgres", "database name")

	sqlImport := flag.String("sql-import", "", "execute sql file and exit")

	flag.Parse()

	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	repo := Repo{}

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

	// create random event
	eventDate := time.Now()
	eventName := eventDate.Format("Mon 2 Jan 15:04:05")
	_, err = repo.insertEvent(Event{Name: eventName, Date: eventDate})
	if err != nil {
		log.Fatalf("cannot create event: %v", err)
	}

	// modify first event with now
	if err := repo.updateEvent(Event{ID: 1, Name: eventName, Date: eventDate}); err != nil {
		log.Fatalf("cannot update event: %v", err)
	}

	// display events
	events, err := repo.getEvents()
	if err != nil {
		log.Fatalf("cannot gather events: %v", err)
	}

	for _, event := range events {
		log.Printf("event %d is: %s", event.ID, event.Name)
	}
}

type Repo struct {
	db *sql.DB
}

// execFile will parse a sql file, extract each query
// by splitting to ";" and execute them onto the database.
// This is not a robust way because it is not a proper sql parser
// but it will be sufficient for the needs of this project.
func (r Repo) execFile(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	reqs := strings.Split(string(f), ";")
	for _, req := range reqs {
		if _, err := r.db.Exec(req); err != nil {
			return err
		}
	}

	return nil
}
