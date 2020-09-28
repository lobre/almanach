package main

import (
	"database/sql"
	"io/ioutil"
	"strings"
)

type repo struct {
	db *sql.DB
}

// execFile will parse a sql file, extract each query
// by splitting to ";" and execute them onto the database.
// This is not a robust way because it is not a proper sql parser
// but it will be sufficient for the needs of this project.
func (r repo) execFile(path string) error {
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
