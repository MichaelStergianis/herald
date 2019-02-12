package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	testfixtures "gopkg.in/testfixtures.v2"
)

var (
	dbName   string
	hdb      *HeraldDB
	fixtures *testfixtures.Context
)

const (
	testLib = "test_lib/"
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

// prepareTestLibrary ...
func prepareTestLibrary() {
	fsPath, err := filepath.Abs("test_lib/")
	check(err)

	err = hdb.AddLibrary("test", fsPath)
	check(err)
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

// TestAddLibrary ...
func TestAddLibrary(t *testing.T) {
	prepareTestDatabase()
	expected := Library{Name: "MusicalTest", Path: "/h/tt/MusicalTest/"}

	err := hdb.AddLibrary(expected.Name, expected.Path)
	if err != nil {
		t.Error(err)
	}

	row := hdb.db.QueryRow("SELECT libraries.name, libraries.fs_path FROM music.libraries WHERE (libraries.name = $1)",
		expected.Name)

	var result Library
	row.Scan(&result.Name, &result.Path)

	if expected != result {
		fmt.Printf("expected: %v\n", expected)
		fmt.Printf("result: %v\n", result)

		t.Error(errors.New("db_test: expected did not equal result"))

	}
}

// TestAddLibraryNoAbs ...
func TestAddLibraryNoAbs(t *testing.T) {
	prepareTestDatabase()
	err := hdb.AddLibrary("NoAbs", "Music/")

	if err != ErrLibAbs {
		t.Error(errors.New("did not get absolute path error"))
	}
}

// TestGetLibraries ...
func TestGetLibraries(t *testing.T) {
	prepareTestDatabase()

	expected := []Library{
		Library{
			ID: 1, Name: "Music", Path: "/home/test/Music/",
		},

		Library{
			ID: 2, Name: "My Music", Path: "/home/tests/MyMusic/",
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
		Artist{ID: 1, Name: "BADBADNOTGOOD", Path: "BADBADNOTGOOD/"},
		Artist{ID: 2, Name: "BADBADNOTGOOD & Ghostface Killah", Path: "BADBADNOTGOOD & Ghostface Killah/"},
		Artist{ID: 3, Name: "Iron Maiden", Path: "Iron Maiden/"},
		Artist{ID: 4, Name: "Megadeth", Path: "Megadeth/"},
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

// TestSongInDatabase ...
func TestSongInDatabase(t *testing.T) {
	prepareTestDatabase()

	var (
		inLib bool
		err   error
	)

	song := Song{
		ID:   3,
		Path: "Iron Maiden/Killers/01 The Ides of March.mp3",
	}

	inLib, err = hdb.songInDatabase(song)

	if err != nil {
		t.Error(err)
	}

	if !inLib {
		t.Error(errors.New("expected song is not in database"))
	}
}

// TestSongNotInDatabase ...
func TestSongNotInDatabase(t *testing.T) {
	prepareTestDatabase()

	var (
		inLib bool
		err   error
	)

	song := Song{
		ID:   9,
		Path: "Iron Maiden/Killers/02 Wrathchild.mp3",
	}

	inLib, err = hdb.songInDatabase(song)

	if err != nil {
		t.Error(err)
	}

	if inLib {
		t.Error(errors.New("unexpected song in database"))
	}
}

// TestGetAlbum ...
func TestGetAlbum(t *testing.T) {
	prepareTestDatabase()

	var (
		a   Album
		err error
	)

	album := Album{
		ID:        1,
		Artist:    1,
		Year:      2011,
		NumTracks: 20,
		NumDisks:  1,
		Title:     "III",
		Path:      "BADBADNOTGOOD/III/",
		Duration:  1688,
	}

	a, err = hdb.GetAlbum(album)

	if err != nil {
		t.Error(err)
	}

	if a != album {
		t.Error(errors.New("expected album is not in database"))
	}
}

// TestGetAlbumNegative ...
func TestGetAlbumNegative(t *testing.T) {
	prepareTestDatabase()

	var (
		a   Album
		err error
	)

	album := Album{
		ID:        9,
		Artist:    1,
		Year:      2011,
		NumTracks: 20,
		Title:     "III",
		Path:      "BADBADNOTGOOD & Tyler the Creator/sessions/",
		Duration:  1688,
	}

	a, err = hdb.GetAlbum(album)

	if err != sql.ErrNoRows && err != nil {
		t.Error(err)
	}

	if a == album {
		t.Error(errors.New("unexpected album is in database"))
	}
}

// TestAddArtist ...
func TestAddArtist(t *testing.T) {
	prepareTestDatabase()

	artist := Artist{
		Name: "Clever Girl",
		Path: "Clever Girl/",
	}

	a, err := hdb.addArtist(artist)

	if err != nil {
		t.Error(err)
	}

	a.ID = 0

	if a != artist {
		t.Error(errors.New("unexpected value returned"))

	}
}

// TestAddAlbum ...
func TestAddAlbum(t *testing.T) {
	prepareTestDatabase()

	album := Album{
		Artist:    4,
		Year:      2001,
		NumTracks: 11,
		NumDisks:  1,
		Title:     "Counterparts",
		Path:      "Rush/Counterparts/",
	}

	a, err := hdb.addAlbum(album)
	if err != nil {
		t.Error(err)
	}

	a.ID = 0

	if a != album {
		t.Error(errors.New("unexpected value returned"))
	}
}

// TestScanLibrary ...
func TestScanLibrary(t *testing.T) {
	// prepare the test library
	prepareTestDatabase()
	prepareTestLibrary()

	artists, err := hdb.GetArtists()
	if err != nil {
		t.Error(err)
	}

	for _, artist := range artists {
		fmt.Printf("%+v\n", artist)
	}

	testLibPath, err := filepath.Abs("test_lib/")
	testLibPath = testLibPath + "/"
	if err != nil {
		t.Error(err)
	}
	libs, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}
	for _, lib := range libs {
		if lib.Path == testLibPath {
			fmt.Printf("%+v\n", lib)
			hdb.ScanLibrary(lib)
		}
	}

	fmt.Println()

	artists, err = hdb.GetArtists()
	if err != nil {
		t.Error(err)
	}

	for _, artist := range artists {
		fmt.Printf("%+v\n", artist)
	}
}
