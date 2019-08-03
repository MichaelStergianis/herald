package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	warblerDB "gitlab.stergianis.ca/michael/warbler/db"
)

const resourcesLoc string = "frontend/resources/public/"

// newServer ...
func newServer(connStr string) (serv *server, err error) {
	serv = &server{}
	serv.wdb, err = warblerDB.Open(connStr)
	if err != nil {
		return &server{}, err
	}

	serv.router = mux.NewRouter()
	return serv, nil
}

// serveMusic ...
func serveMusic(hdb *warblerDB.WarblerDB) func(w http.ResponseWriter, r *http.Request) {
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

	// args
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

	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	serv, err := newServer("host=" + host + " dbname=warbler user=warbler sslmode=disable")
	check(err)
	defer serv.wdb.Close()

	serv.addRoutes()

	log.Fatal(http.ListenAndServe(portString, serv.router))
}
