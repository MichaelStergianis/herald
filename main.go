package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	HeraldDB "gitlab.stergianis.ca/herald/db"
)

const resourcesLoc string = "frontend/resources/public/"

// serveMusic ...
func serveMusic(hdb *HeraldDB.HeraldDB) func(w http.ResponseWriter, r *http.Request) {
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

	hdb := HeraldDB.New()

	count, err := hdb.CountTable("music.libraries")
	check(err)

	fmt.Println("Count is:", count)

	os.Exit(0)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/resources/public/index.html")
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(resourcesLoc, "css")))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(path.Join(resourcesLoc, "img")))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(resourcesLoc, "js")))))
	http.HandleFunc("/stream", serveMusic(hdb))

	err = http.ListenAndServe(portString, nil)
	check(err)
	return
}
