package main

import (
	"database/sql"
	"fmt"

	// pq uses golang sql
	_ "github.com/lib/pq"
)

type library struct {
	name string
	path string
}

type song struct {
}

// createDb ...
// creates the connection to the db
func createDb() *sql.DB {
	connStr := "user=herald dbname=herald_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	check(err)
	return db
}

// initDB ...
// checks whether or not the database exists given the db
func initDB(db *sql.DB) {

}

// createLibrary ...
func createLibrary(db *sql.DB, name string, path string) {
	stmt, err := db.Prepare("INSERT INTO libraries (library_name, fs_path) VALUES ($1, $2);")
	check(err)
	defer stmt.Close()
	res, err := stmt.Exec(name, path)
	check(err)
	fmt.Println(res)
}

// getLibraries ...
func getLibraries(db *sql.DB) []library {
	// query
	rows, err := db.Query(`SELECT library_name, fs_path from libraries;`)
	defer rows.Close()

	var libraries []library
	for rows.Next() {
		var l library
		err = rows.Scan(&l.name, &l.path)
		check(err)
		libraries = append(libraries, l)
	}

	return libraries
}

func scanLibrary(db *sql.DB, lib library) {

}
