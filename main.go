package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "postgres")

	repo := Repo{}

	var err error
	repo.db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("cannot connect to the db: %v", err)
	}
	defer repo.db.Close()

	ss, err := repo.subscriptions(Event{ID: 1})
	if err != nil {
		log.Fatalf("cannot gather subscriptions: %v", err)
	}

	for _, s := range ss {
		spew.Dump(s)
	}
}

type Event struct {
	ID   int
	Name string
	Date time.Time
}

type Subscription struct {
	Subscriber string
	Here       bool
	Comment    string
	EventID    int `db:"event_id"`
}

type Repo struct {
	db *sqlx.DB
}

func (r Repo) events() ([]Event, error) {
	var ee []Event

	if err := r.db.Select(&ee, "SELECT * from events"); err != nil {
		return nil, err
	}

	return ee, nil
}

func (r Repo) subscriptions(e Event) ([]Subscription, error) {
	var ss []Subscription

	stmt, err := r.db.PrepareNamed(`SELECT * FROM subscriptions WHERE event_id = :id`)
	if err != nil {
		return nil, err
	}

	if err := stmt.Select(&ss, e); err != nil {
		return nil, err
	}

	return ss, nil
}
