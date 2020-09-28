package main

import "time"

type Event struct {
	ID   int
	Name string
	Date time.Time
}

func (r Repo) getEvents() ([]Event, error) {
	events := []Event{}

	rows, err := r.db.Query("SELECT id, name, date FROM events order by id")
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

func (r Repo) insertEvent(e Event) (int, error) {
	var id int
	err := r.db.QueryRow("INSERT INTO events(name, date) VALUES($1, $2) RETURNING id", e.Name, e.Date).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r Repo) updateEvent(e Event) error {
	if _, err := r.db.Exec("UPDATE events set name=$1, date=$2 WHERE id=$3", e.Name, e.Date, e.ID); err != nil {
		return err
	}
	return nil
}

func (r Repo) removeEvent(id int) error {
	if _, err := r.db.Exec("DELETE FROM events where id = $1", id); err != nil {
		return err
	}
	return nil
}
