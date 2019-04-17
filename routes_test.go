package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	heraldDB "gitlab.stergianis.ca/michael/herald/db"
	"gopkg.in/testfixtures.v2"
	"olympos.io/encoding/edn"
)

var (
	serv     *server
	fixtures *testfixtures.Context
)

// TestMain ...
func TestMain(m *testing.M) {
	var err error

	// use herald test server
	serv, err = newServer("dbname=herald_test user=herald sslmode=disable")

	if err != nil {
		log.Fatalln("cannot create connection to testing server", err)
	}

	fixtures, err = testfixtures.NewFolder(serv.hdb.DB, &testfixtures.PostgreSQL{UseAlterConstraint: true}, "db/fixtures")
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(m.Run())
}

// TestNewMediaHandler ...
func TestNewMediaHandler(t *testing.T) {

	encoders := [...]encoder{
		{"json", json.Marshal, json.Unmarshal},
		{"edn", edn.Marshal, edn.Unmarshal},
	}
	records := [...]record{
		{"/libraries/", "music.libraries", &heraldDB.Library{ID: 1}},
		{"/genres/", "music.genres", &heraldDB.Genre{ID: 1}},
		{"/artists/", "music.artists", &heraldDB.Artist{ID: 1}},
		{"/albums/", "music.albums", &heraldDB.Album{ID: 1}},
		{"/songs/", "music.songs", &heraldDB.Song{ID: 1}},
		{"/images/", "music.images", &heraldDB.Image{ID: 1}},
	}

	answers := [...]heraldDB.Queryable{
		&heraldDB.Library{ID: 1, Name: "Music"},
		&heraldDB.Genre{ID: 1, Name: "Jazz"},
		&heraldDB.Artist{ID: 1, Name: "BADBADNOTGOOD"},
		&heraldDB.Album{
			ID: 1, Artist: 1, Year: 2011, NumTracks: 20,
			NumDisks: 1, Title: "III", Duration: 1688,
		},
		&heraldDB.Song{
			ID: 1, Album: 1, Genre: 1, Title: "In the Night",
			Track: 1, NumTracks: 20, Disk: 1, NumDisks: 1,
			Size: 204192, Duration: 1993, Artist: "BADBADNOTGOOD",
		},
		&heraldDB.Image{ID: 1},
	}

	// test cases
	for _, enc := range encoders {
		for idx, rec := range records {
			req, err := http.NewRequest("GET", fmt.Sprintf("/?id=%d", rec.query.GetID()), nil)

			if err != nil {
				t.Errorf("could not create request for %s", rec.table)
			}
			rr := httptest.NewRecorder()

			handler := serv.NewUniqueGetHandler(rec.table, enc.name, enc.enc, rec.query)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("handler for %s returned status code %v", rec.url, rr.Code)
			}

			err = enc.dec(rr.Body.Bytes(), rec.query)
			if err != nil {
				t.Errorf("encountered error decoding response: %v", err)
			}

			// reflection required to test that they work
			result := reflect.ValueOf(rec.query).Elem().Interface()
			expected := reflect.ValueOf(answers[idx]).Elem().Interface()
			if result != expected {
				t.Errorf("response did not match expected\n\tresponse: %+v\n\tanswer: %+v",
					rec.query, answers[idx])
			}

		}
	}
}
