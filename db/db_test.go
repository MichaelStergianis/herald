package db

import (
	"errors"
	"log"
	"os"
	"testing"

	testfixtures "gopkg.in/testfixtures.v2"
)

var (
	dbName   string
	hdb      *HeraldDB
	fixtures *testfixtures.Context
)

// paramstoStr ...
// Takes a map of parameters and generates a string.
func paramstoStr(params map[string]string) (s string) {
	s = ""
	for k, v := range params {
		s = s + k + "=" + v + " "
	}
	return s
}

// prepareTestDatabase ...
func prepareTestDatabase() {
	if err := fixtures.Load(); err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error

	hdb, err = Open("dbname=herald_test user=herald sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	fixtures, err = testfixtures.NewFolder(hdb.db, &testfixtures.PostgreSQL{UseAlterConstraint: true}, "fixtures")
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// TestCountTable ...
func TestCountTable(t *testing.T) {
	prepareTestDatabase()

	// positive case
	count, err := hdb.CountTable("music.artists")

	if err != nil {
		t.Error(err)
	}

	if count != 4 {
		t.Fail()
	}

	// negative case
	count, err = hdb.CountTable("music.non-existant")

	if err == nil {
		t.Fail()
	}
}

// TestCreateLibrary ...
func TestCreateLibrary(t *testing.T) {
	prepareTestDatabase()
	expected := Library{Name: "MusicalTest", Path: "/h/tt/MusicalTest"}

	err := hdb.CreateLibrary(expected.Name, expected.Path)
	if err != nil {
		t.Error(err)
	}

	row := hdb.db.QueryRow("SELECT libraries.name, libraries.fs_path FROM music.libraries WHERE (libraries.name = $1)",
		expected.Name)

	var result Library
	row.Scan(&result.Name, &result.Path)

	if expected != result {
		t.Error(errors.New("db_test: expected did not equal result"))

	}
}

// TestGetLibraries ...
func TestGetLibraries(t *testing.T) {
	prepareTestDatabase()

	expected := []Library{
		Library{
			ID: 1, Name: "Music", Path: "/home/test/Music",
		},

		Library{
			ID: 2, Name: "My Music", Path: "/home/tests/MyMusic",
		},
	}

	results, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(results); i++ {
		if expected[i] != results[i] {
			t.Error(errors.New("db_test: expected did not match results"))
		}
	}
}

// TestGetArtists ...
func TestGetArtists(t *testing.T) {
	prepareTestDatabase()

	expected := []Artist{
		Artist{ID: 1, Name: "BADBADNOTGOOD"},
		Artist{ID: 2, Name: "BADBADNOTGOOD & Ghostface Killah"},
		Artist{ID: 3, Name: "Iron Maiden"},
		Artist{ID: 4, Name: "Megadeth"},
	}

	results, err := hdb.GetArtists()
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(results); i++ {
		if expected[i] != results[i] {
			t.Error(errors.New("expected did not match results"))
		}
	}
}
