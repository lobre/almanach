package main

import (
	"database/sql"
	"time"
)

type DB struct {
	*sql.DB
}

func Open(connStr string) (*DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

type Event struct {
	ID   int
	Name string
	Date time.Time
}

func (db *DB) getEvents() ([]Event, error) {
	events := []Event{}

	rows, err := db.Query("SELECT id, name, date FROM events order by id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var date time.Time

		err = rows.Scan(&id, &name, &date)
		if err != nil {
			return events, err
		}

		events = append(events, Event{ID: id, Name: name, Date: date})
	}

	return events, err
}

func (db *DB) insertEvent(e Event) (int, error) {
	var id int
	err := db.QueryRow("INSERT INTO events(name, date) VALUES($1, $2) RETURNING id", e.Name, e.Date).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DB) updateEvent(e Event) error {
	if _, err := db.Exec("UPDATE events set name=$1, date=$2 WHERE id=$3", e.Name, e.Date, e.ID); err != nil {
		return err
	}
	return nil
}

func (db *DB) removeEvent(id int) error {
	if _, err := db.Exec("DELETE FROM events where id = $1", id); err != nil {
		return err
	}
	return nil
}

type Subscription struct {
	Subscriber string
	Here       bool
	Comment    string
}

func (db *DB) getSubscriptions(eventID int) ([]Subscription, error) {
	subs := []Subscription{}

	rows, err := db.Query("SELECT subscriber, here, comment FROM subscriptions WHERE event_id = $1", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sub string
		var here bool
		var comment string

		err = rows.Scan(&sub, &here, &comment)
		if err != nil {
			return subs, err
		}

		subs = append(subs, Subscription{Subscriber: sub, Here: here, Comment: comment})
	}

	return subs, err
}

func (db *DB) insertSubscription(eventID int, sub Subscription) error {
	_, err := db.Exec("INSERT INTO subscriptions(event_id, subscriber, here, comment)"+
		"VALUES($1, $2, $3, $4)", eventID, sub.Subscriber, sub.Here, sub.Comment)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) updateSubscription(eventID int, sub Subscription) error {
	_, err := db.Exec("UPDATE subscriptions set here = $1, comment = $2"+
		"WHERE event_id = $3 AND subscriber = $4", sub.Here, sub.Comment, eventID, sub.Subscriber)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) removeSubscription(eventID int, sub string) error {
	_, err := db.Exec("DELETE FROM subscription where event_id = $1 AND subscriber = $2", eventID, sub)
	if err != nil {
		return err
	}
	return nil
}
