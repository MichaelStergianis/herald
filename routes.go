package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	heraldDB "gitlab.stergianis.ca/michael/herald/db"
	"olympos.io/encoding/edn"
)

type record struct {
	url   string
	table string
	query heraldDB.Queryable
}

type encoder struct {
	name string
	enc  func(interface{}) ([]byte, error)
	dec  func([]byte, interface{}) error
}

// badRequestErr ...
func badRequestErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}

// routes ...
func (serv *server) addRoutes() *server {
	// static
	serv.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/resources/public/index.html")
	})
	serv.router.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(resourcesLoc, "css")))))
	serv.router.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(path.Join(resourcesLoc, "img")))))
	serv.router.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(resourcesLoc, "js")))))

	encoders := []encoder{
		{"edn", edn.Marshal, edn.Unmarshal},
		{"json", json.Marshal, json.Unmarshal},
	}

	records := []record{
		{"/library/", "music.libraries", &heraldDB.Library{}},
		{"/artist/", "music.artists", &heraldDB.Artist{}},
		{"/album/", "music.albums", &heraldDB.Album{}},
		{"/genre/", "music.genres", &heraldDB.Genre{}},
		{"/song/", "music.songs", &heraldDB.Song{}},
		{"/image/", "music.images", &heraldDB.Image{}},
	}

	for _, enc := range encoders {
		subrouter := serv.router.PathPrefix("/" + enc.name + "/").Subrouter()
		for _, rec := range records {
			// add the record type to the subrouter
			subrouter.
				HandleFunc(rec.url, serv.NewUniqueGetHandler(rec.table, enc.name, enc.enc, rec.query)).
				Methods("GET")
		}
	}

	return serv
}

// NewArtistHandler ...
func (serv *server) NewArtistHandler(encoder func(interface{}) ([]byte, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idS, ok := params["id"]
		if !ok {
			badRequestErr(w, errors.New("No ID supplied"))
			return
		}

		id, err := strconv.Atoi(idS)
		if err != nil {
			badRequestErr(w, err)
			return
		}
		a := heraldDB.Artist{
			ID: int64(id),
		}

		a, err = serv.hdb.GetUniqueArtist(a)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		response, err := encoder(a)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		w.Write(response)
	}
}

// NewUniqueGetHandler ...
// Expects a database object, a table name, and a type to use.
func (serv *server) NewUniqueGetHandler(tableName string, encStr string, encoder func(interface{}) ([]byte, error), template heraldDB.Queryable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := heraldDB.NewFromQueryable(template)

		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			badRequestErr(w, err)
			return
		}

		query.SetID(int64(id))
		err = serv.hdb.GetUniqueItem(tableName, query)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		response, err := encoder(query)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", fmt.Sprintf("application/%s", encStr))
		w.Write(response)
	}
}
