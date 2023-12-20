package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	dsn := "host=localhost port=5432 user=postgres password=changeit dbname=wonderland sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
