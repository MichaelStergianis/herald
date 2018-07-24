package main

import "database/sql"

func createDb(dbFile string) *sql.DB {
	db, err := sql.Open("sqlite3", dbFile)
	check(err)
	return db
}
