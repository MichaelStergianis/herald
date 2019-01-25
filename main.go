package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const resourcesLoc string = "frontend/resources/public/"

// serveMusic ...
func serveMusic(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Query()
		// error checking on args is somewhat important

		fmt.Println(args)

	}
	return f
}

func main() {
	port := flag.Int("port", 8080, "The port on which to bind the server")
	flag.Parse()
	portString := ":" + strconv.Itoa(*port)

	db := createDb()

	rows, err := db.Query("select * from songs;")
	check(err)
	defer rows.Close()

	fmt.Println(rows.Err())

	if rows.NextResultSet() {
		fmt.Println("More results")
	}

	for rows.Next() {
		var s song
		var id int
		var albumId int
		var path string
		rows.Scan(&id, &albumId, &path, &s.title)
		fmt.Println()
	}

	fmt.Println(rows.Err())

	os.Exit(0)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/resources/public/index.html")
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(resourcesLoc, "css")))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(path.Join(resourcesLoc, "img")))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(resourcesLoc, "js")))))
	http.HandleFunc("/stream", serveMusic(db))

	err = http.ListenAndServe(portString, nil)
	check(err)
	return
}
