package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	heraldDB "gitlab.stergianis.ca/herald/db"
)

// badRequestErr ...
func badRequestErr(w http.ResponseWriter, err error) {
	w.WriteHeader(400)
	w.Write([]byte(err.Error()))
}

// herald.com/artists/{id}

// NewArtistHandler ...
func (s *server) NewArtistHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idS, ok := params["id"]
		if !ok {
			badRequestErr(w, errors.New("No ID supplied"))
		}

		id, err := strconv.Atoi(idS)
		if err != nil {
			badRequestErr(w, err)
		}
		a := heraldDB.Artist{
			ID: int64(id),
		}

		a, err = s.hdb.GetUniqueArtist(a)
		if err != nil {
			badRequestErr(w, err)
		}

		response, err := json.Marshal(a)
		if err != nil {
			badRequestErr(w, err)
		}

		w.WriteHeader(200)
		w.Write(response)
	}
}

// NewMediaHandler ...
// Expects a database object, a table name, and a type to use.
func (s *server) NewMediaHandler(tableName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idS, ok := params["id"]
		if !ok {
			badRequestErr(w, errors.New("No ID supplied"))
		}

		id, err := strconv.Atoi(idS)
		if err != nil {
			badRequestErr(w, err)
		}
	}
}
