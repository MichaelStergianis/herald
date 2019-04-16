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

type template struct {
	url    string
	table  string
	query  heraldDB.Queryable
	answer heraldDB.Queryable
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

	// encoders := []encoder{
	// 	{"/edn/", edn.Marshal, edn.Unmarshal},
	// 	{"/json/", json.Marshal, json.Unmarshal},
	// }

	// templates := []template{
	// 	{},
	// }

	// for _, enc := range encoders {
	// 	subrouter := serv.router.PathPrefix(enc.name).Subrouter()
	// }

	// edn
	ednEncoder := edn.Marshal
	album := &heraldDB.Album{}
	edn := serv.router.PathPrefix("/edn/").Subrouter()
	edn.Handle("/artist/", serv.NewArtistHandler(ednEncoder))
	edn.Handle("/album/", serv.NewMediaHandler("music.albums", "edn", ednEncoder, album))

	// json
	jsonEncoder := json.Marshal
	json := serv.router.PathPrefix("/json/").Subrouter()
	json.Handle("/artist/", serv.NewArtistHandler(jsonEncoder))
	json.Handle("/album/", serv.NewMediaHandler("music.albums", "json", jsonEncoder, album))

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

// NewMediaHandler ...
// Expects a database object, a table name, and a type to use.
func (serv *server) NewMediaHandler(tableName string, encStr string, encoder func(interface{}) ([]byte, error), template heraldDB.Queryable) http.HandlerFunc {
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
