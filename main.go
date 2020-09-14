package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type config struct {
	dbhost string
	dbport int
	dbuser string
	dbpass string
	dbname string
}

schema := `CREATE TABLE events (
	name text,
	description text);`

func main() {
	conf := config{
		dbhost: "localhost",
		dbport: 5432,
		dbuser: "postgres",
		dbpass: "pass",
		dbname: "postgres",
	}

	dataSrc := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		conf.dbhost, conf.dbport, conf.dbuser, conf.dbpass, conf.dbname)

	var db *sqlx.DB
	db, err := sqlx.Open("postgres", dataSrc)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("db connection success")
}
