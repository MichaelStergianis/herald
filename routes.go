// routes.go
//
// This file describes the routes that the server supports. Warbler
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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	warblerDB "gitlab.stergianis.ca/michael/warbler/db"
	"olympos.io/encoding/edn"
)

// TODO: make error handling more stylistically consistent.

var (
	unixEpoch = time.Unix(0, 0)
)

type server struct {
	wdb    *warblerDB.WarblerDB
	router *mux.Router
}

type encFunc func(interface{}) ([]byte, error)

type record struct {
	url   string
	query warblerDB.Queryable
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

// internalServerError ...
func internalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, "an internal server error occurred")
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
		{"/library", &warblerDB.Library{}},
		{"/artist", &warblerDB.Artist{}},
		{"/album", &warblerDB.Album{}},
		{"/genre", &warblerDB.Genre{}},
		{"/song", &warblerDB.Song{}},
		{"/image", &warblerDB.Image{}},
	}

	for _, enc := range encoders {
		subrouter := serv.router.PathPrefix("/" + enc.name + "/").Subrouter()

		// create libraries
		subrouter.
			PathPrefix("/library").
			Methods(http.MethodPost).
			HandlerFunc(serv.newLibraryCreator(enc))
		subrouter.
			PathPrefix("/scanLibrary/{id}").
			Methods(http.MethodPost).
			HandlerFunc(serv.newLibraryScanner(enc))
		subrouter.
			PathPrefix("/library").
			Methods(http.MethodPut).
			HandlerFunc(serv.newLibraryUpdater(enc))

		// streaming
		subrouter.
			PathPrefix("/stream/{id}").
			Methods(http.MethodGet).
			HandlerFunc(serv.newStreamRoute(enc))

		// echo is disabled in code by default for now, maybe a config
		// option later
		/* subrouter.
		PathPrefix("/echo").
		HandlerFunc(serv.newEchoRoute(enc)) */

		for _, rec := range records {
			// add the record type to the subrouter
			subrouter.
				HandleFunc(rec.url+"/{id}", serv.NewUniqueQueryHandler(enc, rec.query)).
				Methods(http.MethodGet)

			// non unique
			subrouter.
				HandleFunc(rec.url, serv.NewQueryHandler(enc, rec.query)).
				Methods(http.MethodGet)
		}
	}

	return serv
}

// NewUniqueQueryHandler ...
// Expects a database object, a table name, and a type to use.
func (serv *server) NewUniqueQueryHandler(enc encoder, queryType warblerDB.Queryable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// we need a new memory address for our query because we will write an id.
		query := warblerDB.NewFromQueryable(queryType)

		params := mux.Vars(r)

		sID := params["id"]

		id, err := strconv.Atoi(sID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query.SetID(int64(id))
		err = serv.wdb.ReadUnique(query)
		if err == warblerDB.ErrNotPresent {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		response, err := enc.enc(query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
func (serv *server) NewQueryHandler(enc encoder, queryType interface{}) http.HandlerFunc {
	const orderField = "orderby"
	validFields, err := warblerDB.ValidFields(enc.name, queryType)
	if err != nil {
		log.Panicln("Error creating new query handler:", err)
	}
	validFields[orderField] = struct{}{}

	converter := warblerDB.NewTagConverter(queryType, enc.name, "sql")

	return func(w http.ResponseWriter, r *http.Request) {
		query := warblerDB.NewFromInterface(queryType)
		data, ok := r.URL.Query()["data"]
		// if data is not given, return all articles matching that data type
		if !ok {
			data = []string{`{}`}
		}

		var results []interface{}

		var orderBy []string = r.URL.Query()[orderField]

		for _, d := range data {
			// construct the query
			err := enc.dec([]byte(d), query)
			if err != nil {
				badRequestErr(w, err)
				return
			}

			convTags, err := warblerDB.ConvertTags(orderBy, converter)
			if err != nil {
				badRequestErr(w, err)
				return
			}

			result, err := serv.wdb.Read(query, convTags)
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

// newLibraryCreator creates a library creator based on the current
// server. It allows you to issue a request via http to create a
// database object representing the root location of a collection of
// audio files.
func (serv *server) newLibraryCreator(enc encoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			internalServerError(w)
			return
		}

		var l warblerDB.Library
		err = enc.dec(data, &l)
		if err != nil {
			internalServerError(w)
			return
		}

		err = serv.wdb.Create(&l, []string{"id"})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Library already exists.")
			return
		}

		returnData, err := enc.enc(l)
		if err != nil {
			internalServerError(w)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/"+enc.name)
		fmt.Fprintf(w, "%s", returnData)
	}
}

// newLibraryUpdater creates a library updater
func (serv *server) newLibraryUpdater(enc encoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			internalServerError(w)
			return
		}

		var (
			tmp   warblerDB.Library
			set   warblerDB.Library
			where warblerDB.Library
		)

		// read the provided library
		err = enc.dec(data, &tmp)
		if err != nil {
			internalServerError(w)
			return
		}

		// name and path will be updated, any elements not provided in
		// the body will be encoded as zero values and will be ignored
		// by wdb.Update.
		set.Name = tmp.Name
		set.Path = tmp.Path

		// using id to identify where
		where.ID = tmp.ID

		rowsAffected, err := serv.wdb.Update(set, where)
		if err != nil {
			internalServerError(w)
			return
		}

		if rowsAffected != 1 {
			internalServerError(w)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// newLibraryScanner ...
func (serv *server) newLibraryScanner(enc encoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		libID := params["id"]
		id, err := strconv.ParseInt(libID, 10, 64)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		lib := warblerDB.Library{ID: id}
		err = serv.wdb.ReadUnique(&lib)
		if err != nil {
			badRequestErr(w, err)
			return
		}

		err = serv.wdb.ScanLibrary(lib)
		if err != nil {
			internalServerError(w)
			return
		}

		returnData, err := enc.enc(lib)
		if err != nil {
			internalServerError(w)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/"+enc.name)
		fmt.Fprintf(w, "%s", returnData)
	}
}

// newEchoRoute is a helper route that will print the body of any
// request. Can be used to inspect frontend ajax calls.
func (serv *server) newEchoRoute(enc encoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			internalServerError(w)
			return
		}
		fmt.Printf("%s\n", body)
	}
}

// newStreamRoute ...
func (serv *server) newStreamRoute(enc encoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read the song id from the request
		params := mux.Vars(r)
		idS, ok := params["id"]
		if !ok {
			badRequestErr(w, errors.New("no id provided"))
			return
		}

		id, err := strconv.ParseInt(idS, 10, 64)
		if err != nil {
			badRequestErr(w, errors.New("invalid id"))
			return
		}

		song := warblerDB.Song{ID: id}
		err = serv.wdb.ReadUnique(&song)
		if err != nil {
			badRequestErr(w, errors.New("bad lookup"))
			return
		}

		// lookup file for the song
		f, err := os.OpenFile(song.Path, os.O_RDONLY, 0644)
		defer f.Close()
		if err != nil {
			internalServerError(w)
			return
		}

		// serve content supports ranged headers
		http.ServeContent(w, r, song.Path, unixEpoch, f)
	}
}
