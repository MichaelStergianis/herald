package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"
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
	logfile := *flag.String("logfile", "", "The log file to use. Defaults to stdout.")
	flag.Parse()
	portString := ":" + strconv.Itoa(*port)
	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// quit
			fmt.Fprintln(os.Stderr, "Error creating logfile.")
			os.Exit(1)
		}
		log.SetOutput(f)
	}

	hdb, err := heraldDB.Open("dbname=herald user=herald sslmode=disable")
	check(err)

	defer hdb.Close()

	// rest
	http.HandleFunc("/stream", serveMusic(hdb))

	router := mux.NewRouter()

	// static
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/resources/public/index.html")
	})
	router.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(resourcesLoc, "css")))))
	router.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(path.Join(resourcesLoc, "img")))))
	router.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(resourcesLoc, "js")))))

	err = http.ListenAndServe(portString, router)
	check(err)
	return
}
