package main

import (
	"flag"
	"fmt"
	"net/http"
	"path"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	heraldDB "gitlab.stergianis.ca/herald/db"
)

const resourcesLoc string = "frontend/resources/public/"

// serveMusic ...
func serveMusic(hdb *heraldDB.HeraldDB) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Query()
		// error checking on args is somewhat important

		fmt.Println(args)

	}
	return f
}

func main() {
	var err error
	port := flag.Int("port", 8080, "The port on which to bind the server")
	flag.Parse()
	portString := ":" + strconv.Itoa(*port)

	hdb, err := heraldDB.Open("dbname=herald user=herald sslmode=disable")
	check(err)

	defer hdb.Close()

	// static
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/resources/public/index.html")
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(resourcesLoc, "css")))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(path.Join(resourcesLoc, "img")))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(resourcesLoc, "js")))))

	// rest
	http.HandleFunc("/stream", serveMusic(hdb))

	err = http.ListenAndServe(portString, nil)
	check(err)
	return
}
