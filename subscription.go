package main

type Subscription struct {
	Subscriber string
	Here       bool
	Comment    string
}

func (r Repo) getSubscriptions(eventID int) ([]Subscription, error) {
	subs := []Subscription{}

	rows, err := r.db.Query("SELECT subscriber, here, comment FROM subscriptions WHERE event_id = $1", eventID)
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

func (r Repo) insertSubscription(eventID int, sub Subscription) error {
	_, err := r.db.Exec("INSERT INTO subscriptions(event_id, subscriber, here, comment)"+
		"VALUES($1, $2, $3, $4)", eventID, sub.Subscriber, sub.Here, sub.Comment)
	if err != nil {
		return err
	}
	return nil
}

func (r Repo) updateSubscription(eventID int, sub Subscription) error {
	_, err := r.db.Exec("UPDATE subscriptions set here = $1, comment = $2"+
		"WHERE event_id = $3 AND subscriber = $4", sub.Here, sub.Comment, eventID, sub.Subscriber)
	if err != nil {
		return err
	}
	return nil
}

func (r Repo) removeSubscription(eventID int, sub string) error {
	_, err := r.db.Exec("DELETE FROM subscription where event_id = $1 AND subscriber = $2", eventID, sub)
	if err != nil {
		return err
	}
	return nil
}
