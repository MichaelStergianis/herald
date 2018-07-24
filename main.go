package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
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
	dbName := flag.String("dbFile", "./db/library.db", "The db file to use")
	flag.Parse()
	portString := ":" + strconv.Itoa(*port)

	db := createDb(*dbName)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/resources/public/index.html")
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(resourcesLoc+"css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(resourcesLoc+"img"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(resourcesLoc+"js"))))
	http.HandleFunc("/stream", serveMusic(db))

	err := http.ListenAndServe(portString, nil)
	check(err)
	return
}
