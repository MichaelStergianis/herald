package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	edn "olympos.io/encoding/edn"
)

const resourcesLoc string = "frontend/resources/public/"

// Message ...
// message passing struct
type Message struct {
	Type    string `edn:"type"`
	Message string `edn:"message"`
}

// check ...
// Checks errors and upon error exits
func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}

// serveMusic ...
func serveMusic(db *sql.DB, upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != "http://"+r.Host {
			http.Error(w, "Cross Origin not allowed", 403)
			fmt.Println("Cross Origin not allowed")
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		check(err)

		go func(conn *websocket.Conn) {
			defer conn.Close()
			for {
				mType, msg, err := conn.ReadMessage()
				var message Message
				edn.Unmarshall(*msg, &message)
				fmt.Printf("%d: %s\n", mType, string(msg))

				check(err)
			}
		}(conn)
	}
	return f
}

func createDb(dbFile string) *sql.DB {
	db, err := sql.Open("sqlite3", dbFile)
	check(err)
	return db
}

func main() {
	port := flag.Int("port", 8080, "The port on which to bind the server")
	dbName := flag.String("dbFile", "./db/library.db", "The db file to use")
	flag.Parse()
	portString := ":" + strconv.Itoa(*port)

	db := createDb(*dbName)
	upgrader := websocket.Upgrader{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/resources/public/index.html")
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(resourcesLoc+"css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(resourcesLoc+"img"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(resourcesLoc+"js"))))
	http.HandleFunc("/ws", serveMusic(db, upgrader))

	err := http.ListenAndServe(portString, nil)
	check(err)
	return
}
