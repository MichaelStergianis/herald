package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	hdb       *HeraldDB
	prepareDB func()
)

const (
	testLibName = "Test"
	dbName      = "herald_test"
	testSongLoc = "Simpsons/Thermo/"
	testSong    = "Simpsons/Thermo/01 Obey.mp3"
)

var (
	testLib, _ = filepath.Abs("test_lib")
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

// fmtTestLib ...
func fmtTestLib() string {
	fsPath, err := filepath.Abs(testLib)
	check(err)
	return fsPath
}

// prepareTestLibrary ...
func prepareTestLibrary() {
	fsPath := fmtTestLib()

	err := hdb.AddLibrary(testLibName, fsPath)
	check(err)
}

func TestMain(m *testing.M) {
	var err error
	hdb, err = Open("dbname=" + dbName + " user=herald sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	prepareDB, err = PrepareTestDatabase(hdb, "fixtures")
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// TestStripToArtist ...
func TestStripToArtist(t *testing.T) {
	song := Song{
		ID:   3,
		Path: "/home/tests/MyMusic/Iron Maiden/Killers/01 The Ides of March.mp3",
	}

	lib := Library{
		Path: "/home/tests/MyMusic",
	}

	fsPath := stripToArtist(song.Path, lib)

	if fsPath != "/home/tests/MyMusic/Iron Maiden" {
		t.Error("incorrect string returned from strip")
	}
}

// TestStripToAlbum ...
func TestStripToAlbum(t *testing.T) {
	song := Song{
		ID:   3,
		Path: "/home/tests/MyMusic/Iron Maiden/Killers/01 The Ides of March.mp3",
	}

	artist := Artist{
		Path: "/home/tests/MyMusic/Iron Maiden",
	}

	fsPath := stripToAlbum(song.Path, artist)

	if fsPath != "/home/tests/MyMusic/Iron Maiden/Killers" {
		t.Error("incorrect string returned from strip")
	}
}

// TestCountTable ...
func TestCountTable(t *testing.T) {
	prepareDB()

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
	prepareDB()
	expected := Library{Name: "MusicalTest", Path: "/h/tt/MusicalTest/"}

	err := hdb.AddLibrary(expected.Name, expected.Path)
	if err != nil {
		t.Error(err)
	}

	row := hdb.QueryRow("SELECT libraries.name, libraries.fs_path FROM music.libraries WHERE (libraries.name = $1)",
		expected.Name)

	var result Library
	row.Scan(&result.Name, &result.Path)

	if expected != result {
		t.Error("expected did not equal result")

	}
}

// TestAddLibraryNoAbs ...
func TestAddLibraryNoAbs(t *testing.T) {
	prepareDB()
	err := hdb.AddLibrary("NoAbs", "Music/")

	if err != ErrNotAbs {
		t.Error("did not get absolute path error")
	}
}

// TestGetLibraries ...
func TestGetLibraries(t *testing.T) {
	prepareDB()

	expected := map[string]Library{
		"Music": Library{
			ID: 1, Name: "Music", Path: "/home/test/Music",
		},

		"My Music": Library{
			ID: 2, Name: "My Music", Path: "/home/tests/MyMusic",
		},
	}

	results, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	for name, lib := range results {
		if expected[name] != lib {
			t.Error("expected did not match results")
		}
	}
}

// TestSongInLibrary ...
func TestSongInLibrary(t *testing.T) {
	prepareDB()

	var (
		inLib bool
		err   error
	)

	libs, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	song := Song{
		ID:   3,
		Path: "/home/tests/MyMusic/Iron Maiden/Killers/01 The Ides of March.mp3",
	}

	inLib, err = hdb.songInLibrary(song, libs["My Music"])

	if err != nil {
		t.Error(err)
	}

	if !inLib {
		t.Error("expected song is not in database")
	}
}

// TestSongNotInLibrary ...
func TestSongNotInLibrary(t *testing.T) {
	prepareDB()

	var (
		inLib bool
		err   error
	)

	libs, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	song := Song{
		Path: "Iron Maiden/Killers/02 Wrathchild.mp3",
	}

	inLib, err = hdb.songInLibrary(song, libs["My Music"])

	if err != nil {
		t.Error(err)
	}

	if inLib {
		t.Error(errors.New("unexpected song in database"))
	}
}

// TestDuration ...
func TestDuration(t *testing.T) {
	prepareDB()
	prepareTestLibrary()

	expectedDuration := 3.46

	song := Song{
		Path: path.Join(testLib, testSong),
	}

	d, err := duration(song)

	if err != nil {
		t.Error(err)
	}

	if d != expectedDuration {
		t.Error("unexpected return value")
	}
}

// TestAddSongToLibrary ...
func TestAddSongToLibrary(t *testing.T) {
	prepareDB()
	prepareTestLibrary()

	// testSong := path.Join(testLib, "GoldLink/At What Cost/02 Same Clothes as Yesterday.m4a")
}

// TestProcessMedia ...
func TestProcessMedia(t *testing.T) {
	prepareDB()
	prepareTestLibrary()

	libs, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	ts := path.Join(testLib, testSong)
	fmt.Printf("%s\n", ts)

	err = hdb.processMedia(ts, libs[testLibName])
	if err != nil {
		t.Error(err)
	}
}

// TestScanLibrary ...
func TestScanLibrary(t *testing.T) {
	prepareDB()
	prepareTestLibrary()

	libs, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	err = hdb.ScanLibrary(libs[testLibName])
	if err != nil {
		t.Error(err)
	}

	songs, err := hdb.GetSongsInLibrary(libs[testLibName])
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%v\n", songs)

	for _, song := range songs {
		fmt.Printf("%+v\n", song)
	}
}

// TestQuerySelection ...
func TestQuerySelection(t *testing.T) {
	testCases := []struct {
		rQuery     reflect.Value
		query      string
		values     []interface{}
		nullValues []sql.Scanner
		err        error
	}{
		{reflect.ValueOf(&Artist{}), "SELECT id, name, fs_path",
			[]interface{}{
				new(int64),
				new(string),
				new(string),
			},
			[]sql.Scanner{
				&sql.NullInt64{},
				&sql.NullString{},
				&sql.NullString{},
			}, nil},

		// give a non pointer, get an error
		{reflect.ValueOf(Genre{}), "", nil, nil, ErrCannotAddr},
	}

	for testCase, test := range testCases {
		q, v, nV, err := querySelection(test.rQuery)
		// query string check
		if q != test.query {
			t.Errorf("returned query did not match expected during test case: %d\n"+
				"\t\texpected: %v\n"+
				"\t\tresult:   %v\n", testCase, q, test.query)
		}

		// values check
		for i := 0; i < len(v); i++ {
			result := reflect.ValueOf(v[i])
			expected := reflect.ValueOf(test.values[i])
			if result.Elem().Interface() != expected.Elem().Interface() {
				t.Errorf("returned value did not match expected during test case %d\n"+
					"\t\texpected: %v\n"+
					"\t\tresult:   %v\n", testCase, expected.Elem(), result.Elem())
			}
		}

		// null values check
		for i := 0; i < len(v); i++ {
			result := reflect.ValueOf(nV[i])
			expected := reflect.ValueOf(test.nullValues[i])
			if result.Elem().Interface() != expected.Elem().Interface() {
				t.Errorf("returned value did not match expected during test case: %d\n"+
					"\t\texpected: %v\n"+
					"\t\tresult:   %v\n", testCase, expected.Elem(), result.Elem())
			}
		}

		// error check
		if err != test.err {
			t.Error(err)
		}

	}
}

// TestGetUniqueItem ...
func TestGetUniqueItem(t *testing.T) {
	prepareDB()

	query := &Artist{
		ID: 1,
	}

	verify := Artist{
		ID:   1,
		Name: "BADBADNOTGOOD",
		Path: "/home/test/Music/BADBADNOTGOOD",
	}

	err := hdb.GetUniqueItem(query)
	if err != nil {
		t.Error(err)
	}

	if verify != *query {
		t.Error("result did not match expected")
	}

}

// TestGetItem ...
func TestGetItem(t *testing.T) {
	prepareDB()

	testCases := [...]struct {
		query   interface{}
		encName string
		orderby []string
		answer  []interface{}
	}{
		{Song{Artist: "BADBADNOTGOOD"}, "edn", []string{},
			[]interface{}{
				Song{ID: 1, Album: 1, Genre: 1,
					Path:  "/home/test/Music/BADBADNOTGOOD/III/01 In the Night.mp3",
					Title: "In the Night",
					Track: 1, NumTracks: 20, Disk: 1, NumDisks: 1,
					Size: 204192, Duration: 1993,
					Artist: "BADBADNOTGOOD"},
				Song{ID: 5, Album: 1, Genre: 1,
					Path:  "/home/test/Music/BADBADNOTGOOD/III/02 Triangle.mp3",
					Title: "Triangle",
					Track: 2, NumTracks: 20, Disk: 1, NumDisks: 1,
					Size: 204299, Duration: 1999,
					Artist: "BADBADNOTGOOD"}}},
		// order by single element
		{Album{Artist: 1}, "json", []string{"num-tracks"},
			[]interface{}{
				Album{ID: 5, Artist: 1, Year: 2012,
					NumTracks: 19, NumDisks: 1, Duration: 1688,
					Title: "IV", Path: "/home/test/Music/BADBADNOTGOOD/IV"},
				Album{ID: 1, Artist: 1, Year: 2011,
					NumTracks: 20, NumDisks: 1, Duration: 1688,
					Title: "III", Path: "/home/test/Music/BADBADNOTGOOD/III"},
			},
		},

		// order by multiple elemnts
		{Album{NumDisks: 1}, "edn", []string{"duration", "num-tracks"},
			[]interface{}{
				Album{ID: 5, Artist: 1, Year: 2012,
					NumTracks: 19, NumDisks: 1, Duration: 1688,
					Title: "IV", Path: "/home/test/Music/BADBADNOTGOOD/IV"},
				Album{ID: 1, Artist: 1, Year: 2011,
					NumTracks: 20, NumDisks: 1, Duration: 1688,
					Title: "III", Path: "/home/test/Music/BADBADNOTGOOD/III"},
				Album{ID: 4, Artist: 4, Year: 1985,
					NumTracks: 13, NumDisks: 1, Duration: 1756,
					Title: "Rust in Peace", Path: "/home/test/Music/Megadeth/Rust in Peace"},
				Album{ID: 2, Artist: 2, Year: 2001,
					NumTracks: 10, NumDisks: 1, Duration: 1800,
					Title: "Sour Soul", Path: "/home/test/Music/BADBADNOTGOOD & Ghostface Killah/Sour Soul"},
				Album{ID: 3, Artist: 3, Year: 1980,
					NumTracks: 8, NumDisks: 1, Duration: 15440,
					Title: "Killers", Path: "/home/tests/MyMusic/Iron Maiden/Killers"},
			}},
		{Album{NumDisks: 1}, "edn", []string{"duration", "year"},
			[]interface{}{
				Album{ID: 1, Artist: 1, Year: 2011,
					NumTracks: 20, NumDisks: 1, Duration: 1688,
					Title: "III", Path: "/home/test/Music/BADBADNOTGOOD/III"},
				Album{ID: 5, Artist: 1, Year: 2012,
					NumTracks: 19, NumDisks: 1, Duration: 1688,
					Title: "IV", Path: "/home/test/Music/BADBADNOTGOOD/IV"},
				Album{ID: 4, Artist: 4, Year: 1985,
					NumTracks: 13, NumDisks: 1, Duration: 1756,
					Title: "Rust in Peace", Path: "/home/test/Music/Megadeth/Rust in Peace"},
				Album{ID: 2, Artist: 2, Year: 2001,
					NumTracks: 10, NumDisks: 1, Duration: 1800,
					Title: "Sour Soul", Path: "/home/test/Music/BADBADNOTGOOD & Ghostface Killah/Sour Soul"},
				Album{ID: 3, Artist: 3, Year: 1980,
					NumTracks: 8, NumDisks: 1, Duration: 15440,
					Title: "Killers", Path: "/home/tests/MyMusic/Iron Maiden/Killers"},
			}},
	}

	for testCase, test := range testCases {
		converter := NewTagConverter(test.query, test.encName, "sql")
		convTags, err := ConvertTags(test.orderby, converter)
		if err != nil {
			t.Errorf("error in tag conversion, test case: %d: %v\n", testCase, err)
		}
		results, err := hdb.GetItem(test.query, convTags)

		if err != nil {
			t.Error(err)
		}

		for i := range results {
			if test.answer[i] != results[i] {
				t.Error(fmt.Errorf("test case %d failed\n\t%9s %+v\n\t%-9s %+v",
					testCase, "expected:", test.answer[i], "result:", results[i]))
			}
		}
	}
}

// TestAddItem ...
func TestAddItem(t *testing.T) {
	testCases := [...]struct {
		query     interface{}
		returning []string
		answer    interface{}
		expErr    error
	}{
		// no genre
		{
			&Song{
				Album: 1, Path: "/home/test/Music/BADBADNOTGOOD/III/03 Sax Stuff.mp3",
				Title: "Sax Stuff", Track: 3, NumTracks: 20, Disk: 1, NumDisks: 1, Size: 21134,
				Duration: 168, Artist: "BADBADNOTGOOD & LeLand WILLY"},
			[]string{"id"},
			&Song{ID: 10001, Album: 1, Genre: 0,
				Path:  "/home/test/Music/BADBADNOTGOOD/III/03 Sax Stuff.mp3",
				Title: "Sax Stuff", Track: 3, NumTracks: 20, Disk: 1, NumDisks: 1,
				Size: 21134, Duration: 168, Artist: "BADBADNOTGOOD & LeLand WILLY"},
			nil,
		},
		// add a genre, then use it
		{
			&Genre{Name: "Jazz Hop"}, []string{"id"}, &Genre{10001, "Jazz Hop"}, nil,
		},

		{
			&Song{},
			[]string{"id"},
			&Song{},
			ErrNonUnique{},
		},
	}

	for testCase, test := range testCases {
		q, err := hdb.addItem(test.query, test.returning)

		if err != test.expErr {
			t.Errorf("test case %d failed: %v\n", testCase, err)
		}

		// reflection is the only way to compare these as they are pointers
		rQ := reflect.ValueOf(test.query)
		rA := reflect.ValueOf(test.answer)
		if rQ.Elem().Interface() != rA.Elem().Interface() {
			t.Errorf("test case %d failed:\n\texpected: %v\n\tresult:   %v\n", testCase, test.answer, q)
		}
	}
}
