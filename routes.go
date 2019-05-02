// routes.go
//
// This file describes the routes that the server supports. Herald
// currently supports two data formats for rest communication, json
// and edn.
//
// For submitting calls to the rest api, all provided data uses the
// "data" argument name. The data should be of the form of the
// corresponding format.
//
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	heraldDB "gitlab.stergianis.ca/michael/herald/db"
	"olympos.io/encoding/edn"
)

type encFunc func(interface{}) ([]byte, error)

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
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(err.Error()))
}

// routes ...
func (serv *server) addRoutes() *server {
	// static
	serv.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(resourcesLoc, "index.html"))
	})
	serv.router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(resourcesLoc, "css")))))
	serv.router.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir(path.Join(resourcesLoc, "img")))))
	serv.router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(resourcesLoc, "js")))))

	encoders := []encoder{
		{"edn", edn.Marshal, edn.Unmarshal},
		{"json", json.Marshal, json.Unmarshal},
	}

	records := []record{
		{"/library", "music.libraries", &heraldDB.Library{}},
		{"/artist", "music.artists", &heraldDB.Artist{}},
		{"/album", "music.albums", &heraldDB.Album{}},
		{"/genre", "music.genres", &heraldDB.Genre{}},
		{"/song", "music.songs", &heraldDB.Song{}},
		{"/image", "music.images", &heraldDB.Image{}},
	}

	for _, enc := range encoders {
		subrouter := serv.router.PathPrefix("/" + enc.name + "/").Subrouter()
		for _, rec := range records {
			// add the record type to the subrouter
			subrouter.
				HandleFunc(rec.url+"/{id}", serv.NewUniqueQueryHandler(rec.table, enc, rec.query)).
				Methods("GET")

			// non unique
			subrouter.
				HandleFunc(rec.url+"s/", serv.NewQueryHandler(rec.table, enc, rec.query)).
				Methods("GET")
		}
	}

	return serv
}

// NewUniqueQueryHandler ...
// Expects a database object, a table name, and a type to use.
func (serv *server) NewUniqueQueryHandler(tableName string, enc encoder, queryType heraldDB.Queryable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := heraldDB.NewFromQueryable(queryType)

		params := mux.Vars(r)

		sID := params["id"]

		id, err := strconv.Atoi(sID)
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

		response, err := enc.enc(query)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/"+enc.name)
		w.Write(response)
	}
}

// NewQueryHandler ...
// Creates a general purpose query handler. Will always write an array of arrays of values for response.
//
// data    - Format corresponding to the encoder.
// orderby - Specifies the field by which to order the data, and is optional.
func (serv *server) NewQueryHandler(tableName string, enc encoder, queryType interface{}) http.HandlerFunc {
	const orderField = "orderby"
	validFields, err := heraldDB.ValidFields(enc.name, queryType)
	if err != nil {
		log.Panicln("Error creating new query handler:", err)
	}
	validFields[orderField] = struct{}{}

	return func(w http.ResponseWriter, r *http.Request) {
		query := heraldDB.NewFromInterface(queryType)
		data, ok := r.URL.Query()["data"]
		// if data is not given, return all articles matching that data type
		if !ok {
			data = []string{`{}`}
		}

		var results []interface{}

		orderBy, ok := r.URL.Query()["orderby"]
		if len(orderBy) == 0 || !ok {
			orderBy = append(orderBy, "")
		}

		for _, d := range data {
			// construct the query
			err := enc.dec([]byte(d), query)
			if err != nil {
				badRequestErr(w, err)
				return
			}

			result, err := serv.hdb.GetItem(tableName, query, orderBy[0])
			if err != nil {
				badRequestErr(w, err)
				return
			}
			results = append(results, result)
		}
		response, err := enc.enc(results)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/"+enc.name)
		w.Write(response)
	}
}
