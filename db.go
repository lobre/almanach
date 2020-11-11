package main

import (
	"database/sql"
	"os"
	"time"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, bool, error) {
	var isNew bool

	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err := os.Create(path)
		if err != nil {
			return nil, isNew, err
		}
		isNew = true
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, isNew, err
	}

	if err := db.Ping(); err != nil {
		return nil, isNew, err
	}

	return &DB{db}, isNew, nil
}

func (db *DB) CreateSchema() error {
	sql := `CREATE TABLE IF NOT EXISTS events (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    date INTEGER NOT NULL,
	comment TEXT
	);`

	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS subscriptions (
    event_id INTEGER NOT NULL,
    subscriber TEXT NOT NULL,
    here INTEGER NOT NULL,
    comment TEXT,
    PRIMARY KEY (event_id, subscriber),
	FOREIGN KEY (event_id) REFERENCES events(id)
	)`

	stmt, err = db.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}

type Event struct {
	ID      int
	Name    string
	Date    time.Time
	Comment string
}

func (db *DB) getEvents() ([]Event, error) {
	events := []Event{}

	rows, err := db.Query("SELECT id, name, date, comment FROM events order by id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var timestamp int64
		var comment string

		err = rows.Scan(&id, &name, &timestamp, &comment)
		if err != nil {
			return events, err
		}

		events = append(events, Event{
			ID:      id,
			Name:    name,
			Date:    time.Unix(timestamp, 0),
			Comment: comment,
		})
	}

	return events, err
}

func (db *DB) insertEvent(e Event) (int, error) {
	sql := "INSERT INTO events(name, date, comment) VALUES(?, ?, ?)"
	res, err := db.Exec(sql, e.Name, e.Date.Unix(), e.Comment)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return int(id), nil
}

func (db *DB) updateEvent(e Event) error {
	sql := "UPDATE events set name=?, date=?, comment=? WHERE id=?"
	if _, err := db.Exec(sql, e.Name, e.Date.Unix(), e.Comment, e.ID); err != nil {
		return err
	}
	return nil
}

func (db *DB) removeEvent(id int) error {
	sql := "DELETE FROM events where id = ?"
	if _, err := db.Exec(sql, id); err != nil {
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

	rows, err := db.Query("SELECT subscriber, here, comment FROM subscriptions WHERE event_id = ?", eventID)
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
	sql := "INSERT INTO subscriptions(event_id, subscriber, here, comment) VALUES(?, ?, ?, ?)"
	if _, err := db.Exec(sql, eventID, sub.Subscriber, sub.Here, sub.Comment); err != nil {
		return err
	}
	return nil
}

func (db *DB) updateSubscription(eventID int, sub Subscription) error {
	sql := "UPDATE subscriptions set here = ?, comment = ? WHERE event_id = ? AND subscriber = ?"
	if _, err := db.Exec(sql, sub.Here, sub.Comment, eventID, sub.Subscriber); err != nil {
		return err
	}
	return nil
}

func (db *DB) removeSubscription(eventID int, sub string) error {
	sql := "DELETE FROM subscription where event_id = ? AND subscriber = ?"
	if _, err := db.Exec(sql, eventID, sub); err != nil {
		return err
	}
	return nil
}
