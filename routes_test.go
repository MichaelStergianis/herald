package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	warblerDB "gitlab.stergianis.ca/michael/warbler/db"
	"olympos.io/encoding/edn"
)

var (
	serv      *server
	prepareDB func()
)

const (
	dbName = "warbler_test"
)

// TestMain ...
func TestMain(m *testing.M) {
	var err error

	// use warbler test server
	serv, err = newServer("dbname=" + dbName + " user=warbler sslmode=disable")
	if err != nil {
		log.Fatalln("cannot create connection to testing server", err)
	}
	serv.addRoutes()

	prepareDB, err = warblerDB.PrepareTestDatabase(serv.hdb, "db/fixtures")
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// TestNewUniqueQueryHandler ...
func TestNewUniqueQueryHandler(t *testing.T) {
	prepareDB()
	encoders := [...]encoder{
		{"json", json.Marshal, json.Unmarshal},
		{"edn", edn.Marshal, edn.Unmarshal},
	}
	records := [...]record{
		{"library", &warblerDB.Library{ID: 1}},
		{"genre", &warblerDB.Genre{ID: 1}},
		{"artist", &warblerDB.Artist{ID: 1}},
		{"album", &warblerDB.Album{ID: 1}},
		{"song", &warblerDB.Song{ID: 1}},
		{"image", &warblerDB.Image{ID: 1}},
	}

	answers := [...]warblerDB.Queryable{
		&warblerDB.Library{ID: 1, Name: "Music"},
		&warblerDB.Genre{ID: 1, Name: "Jazz"},
		&warblerDB.Artist{ID: 1, Name: "BADBADNOTGOOD"},
		&warblerDB.Album{
			ID: 1, Artist: warblerDB.NewNullInt64(1), Year: warblerDB.NewNullInt64(2011), NumTracks: warblerDB.NewNullInt64(20),
			NumDisks: warblerDB.NewNullInt64(1), Title: "III", Duration: warblerDB.NewNullFloat64(1688),
		},
		&warblerDB.Song{
			ID: 1, Album: warblerDB.NewNullInt64(1), Genre: warblerDB.NewNullInt64(1), Title: "In the Night",
			Track: warblerDB.NewNullInt64(1), NumTracks: warblerDB.NewNullInt64(20),
			Disk: warblerDB.NewNullInt64(1), NumDisks: warblerDB.NewNullInt64(1),
			Size: 204192, Duration: 1993, Artist: warblerDB.NewNullString("BADBADNOTGOOD"),
		},
		&warblerDB.Image{ID: 1},
	}

	// test cases
	for _, enc := range encoders {
		for idx, rec := range records {
			req, err := http.NewRequest("GET", fmt.Sprintf("/%s/%s/%d", enc.name, rec.url, rec.query.GetID()), nil)
			if err != nil {
				t.Errorf("error creating request: %v", err)
			}

			rr := httptest.NewRecorder()

			serv.router.ServeHTTP(rr, req)

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

	// secondary cases (incorrect data)
	urls := []string{
		"/edn/album/h9h",
		"/edn/album/99",
	}

	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Error(err)
		}
		rr := httptest.NewRecorder()
		serv.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("handler for %s returned status code %v", url, rr.Code)
		}
	}
}

// TestNewQueryHandler ...
func TestNewQueryHandler(t *testing.T) {
	prepareDB()

	ednE := encoder{"edn", edn.Marshal, edn.Unmarshal}
	jsonE := encoder{"json", json.Marshal, json.Unmarshal}

	cases := []struct {
		enc    encoder
		status int
		url    string
		data   []string
	}{
		// okay cases
		{ednE, http.StatusOK, "/edn/songs/", []string{`{:artist "BADBADNOTGOOD"}`}},
		{ednE, http.StatusOK, "/edn/songs/", []string{`{:artist "Iron Maiden"}`, `{:artist "Megadeth"}`}},
		{jsonE, http.StatusOK, "/json/albums/", []string{`{"num-disks": 1}`}},
		{ednE, http.StatusOK, "/edn/artists/", []string{`{}`}},
		{jsonE, http.StatusOK, "/json/artists/", []string{`{}`}},
		{jsonE, http.StatusOK, "/json/albums/", []string{}},

		// error cases
		{ednE, http.StatusBadRequest, "/edn/songs/", []string{`4`}},
	}
	answers := []string{
		`[[{:id 1 :album 1 :genre 1 :title"In the Night":track 1 :num-tracks 20 :disk 1 :num-disks 1 :size 204192 :duration 1993.0 :artist"BADBADNOTGOOD"}{:id 5 :album 1 :genre 1 :title"Triangle":track 2 :num-tracks 20 :disk 1 :num-disks 1 :size 204299 :duration 1999.0 :artist"BADBADNOTGOOD"}]]`,
		`[[{:id 3 :album 3 :genre 3 :title"The Ides of March":track 1 :num-tracks 8 :disk 1 :num-disks 1 :size 2109 :duration 210.0 :artist"Iron Maiden"}][{:id 4 :album 4 :genre 4 :title"Hangar 18":track 1 :num-tracks 13 :disk 1 :num-disks 1 :size 99948 :duration 9994.0 :artist"Megadeth"}]]`,
		`[[{"id":1,"artist":1,"year":2011,"num-tracks":20,"num-disks":1,"title":"III","duration":1688},{"id":2,"artist":2,"year":2001,"num-tracks":10,"num-disks":1,"title":"Sour Soul","duration":1800},{"id":3,"artist":3,"year":1980,"num-tracks":8,"num-disks":1,"title":"Killers","duration":15440},{"id":4,"artist":4,"year":1985,"num-tracks":13,"num-disks":1,"title":"Rust in Peace","duration":1756},{"id":5,"artist":1,"year":2012,"num-tracks":19,"num-disks":1,"title":"IV","duration":1688}]]`,
		`[[{:id 1 :name"BADBADNOTGOOD"}{:id 2 :name"BADBADNOTGOOD \u0026 Ghostface Killah"}{:id 3 :name"Iron Maiden"}{:id 4 :name"Megadeth"}]]`,
		`[[{"id":1,"name":"BADBADNOTGOOD"},{"id":2,"name":"BADBADNOTGOOD \u0026 Ghostface Killah"},{"id":3,"name":"Iron Maiden"},{"id":4,"name":"Megadeth"}]]`,
		`[[{"id":1,"artist":1,"year":2011,"num-tracks":20,"num-disks":1,"title":"III","duration":1688},{"id":2,"artist":2,"year":2001,"num-tracks":10,"num-disks":1,"title":"Sour Soul","duration":1800},{"id":3,"artist":3,"year":1980,"num-tracks":8,"num-disks":1,"title":"Killers","duration":15440},{"id":4,"artist":4,"year":1985,"num-tracks":13,"num-disks":1,"title":"Rust in Peace","duration":1756},{"id":5,"artist":1,"year":2012,"num-tracks":19,"num-disks":1,"title":"IV","duration":1688}]]`,
		"edn: cannot unmarshal int into Go value of type db.Song",
	}

	for i, testCase := range cases {
		query := testCase.url
		if len(testCase.data) > 0 {
			query += "?data=" + strings.Join(testCase.data, "&data=")
		}

		req, err := http.NewRequest("GET", query, nil)
		if err != nil {
			t.Errorf("error creating request %v", err)
		}

		rr := httptest.NewRecorder()

		serv.router.ServeHTTP(rr, req)

		if rr.Code != testCase.status {
			t.Errorf("handler for %s returned status code %v", req.URL.String(), rr.Code)
		}

		result := string(rr.Body.Bytes())

		if answers[i] != result {
			t.Errorf("answer %d does not match result\n\tquery: %v\n\tresult: %v\n\tanswer: %v", i+1, query, result, answers[i])
		}
	}

}
