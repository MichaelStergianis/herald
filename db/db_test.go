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
	wdb       *WarblerDB
	prepareDB func()
)

const (
	testLibName = "Test"
	dbName      = "warbler_test"
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

	err := wdb.AddLibrary(testLibName, fsPath)
	check(err)
}

func TestMain(m *testing.M) {
	var err error
	wdb, err = Open("dbname=" + dbName + " user=warbler sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	prepareDB, err = PrepareTestDatabase(wdb, "fixtures")
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// TestCountTable ...
func TestCountTable(t *testing.T) {
	prepareDB()

	// positive case
	count, err := wdb.CountTable("music.artists")

	if err != nil {
		t.Error(err)
	}

	if count != 4 {
		t.Fail()
	}

	// negative case
	count, err = wdb.CountTable("music.non-existant")

	if err == nil {
		t.Fail()
	}
}

// TestAddLibrary ...
func TestAddLibrary(t *testing.T) {
	prepareDB()
	expected := Library{Name: "MusicalTest", Path: "/h/tt/MusicalTest/"}

	err := wdb.AddLibrary(expected.Name, expected.Path)
	if err != nil {
		t.Error(err)
	}

	row := wdb.QueryRow("SELECT libraries.name, libraries.fs_path FROM music.libraries WHERE (libraries.name = $1)",
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
	err := wdb.AddLibrary("NoAbs", "Music/")

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

	results, err := wdb.GetLibraries()
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

	libs, err := wdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	song := Song{
		ID:   3,
		Path: "/home/tests/MyMusic/Iron Maiden/Killers/01 The Ides of March.mp3",
	}

	inLib, err = wdb.songInLibrary(song, libs["My Music"])

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

	libs, err := wdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	song := Song{
		Path: "Iron Maiden/Killers/02 Wrathchild.mp3",
	}

	inLib, err = wdb.songInLibrary(song, libs["My Music"])

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

	libs, err := wdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	ts := path.Join(testLib, testSong)

	err = wdb.processMedia(ts, libs[testLibName])
	if err != nil {
		t.Error(err)
	}
}

// TestScanLibrary ...
func TestScanLibrary(t *testing.T) {
	var numSongsExpected = 1
	prepareDB()
	prepareTestLibrary()

	libs, err := wdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	err = wdb.ScanLibrary(libs[testLibName])
	if err != nil {
		t.Error(err)
	}

	songs, err := wdb.GetSongsInLibrary(libs[testLibName])
	if err != nil {
		t.Error(err)
	}

	if len(songs) != numSongsExpected {
		t.Errorf("unexpected number of songs: %d", len(songs))
	}

	expectedSong := Song{ID: 10001,
		Album: NullInt64{NullInt64: sql.NullInt64{10001, true}},
		Genre: NullInt64{NullInt64: sql.NullInt64{0, false}},
		Path:  path.Join(testLib, testSong), Title: "Obey", Size: 56417, Duration: 3.46,
		Track:     NullInt64{NullInt64: sql.NullInt64{1, true}},
		NumTracks: NullInt64{NullInt64: sql.NullInt64{1, true}},
		Disk:      NullInt64{NullInt64: sql.NullInt64{0, false}},
		NumDisks:  NullInt64{NullInt64: sql.NullInt64{0, false}},
		Artist:    NullString{NullString: sql.NullString{"", false}}}

	if songs[0] != expectedSong {
		t.Errorf("unexpected song parsed\n\texpected: %v\n\tresult: %v\n", expectedSong, songs[0])
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
		{reflect.ValueOf(&Artist{}), "SELECT id, name",
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
		q, v, err := querySelection(test.rQuery)
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
	}

	err := wdb.ReadUnique(query)
	if err != nil {
		t.Error(err)
	}

	if verify != *query {
		t.Error("result did not match expected")
	}

}

// TestRead ...
func TestRead(t *testing.T) {
	testCases := [...]struct {
		name    string
		query   interface{}
		orderby []string
		answer  []interface{}
	}{
		{"lookup songs by song-artist", Song{Artist: NewNullString("BADBADNOTGOOD")}, []string{},
			[]interface{}{
				Song{ID: 1, Album: NewNullInt64(1), Genre: NewNullInt64(1),
					Path:  "/home/test/Music/BADBADNOTGOOD/III/01 In the Night.mp3",
					Title: "In the Night",
					Track: NewNullInt64(1), NumTracks: NewNullInt64(20),
					Disk: NewNullInt64(1), NumDisks: NewNullInt64(1),
					Size: 204192, Duration: 1993,
					Artist: NewNullString("BADBADNOTGOOD")},
				Song{ID: 5, Album: NewNullInt64(1), Genre: NewNullInt64(1),
					Path:  "/home/test/Music/BADBADNOTGOOD/III/02 Triangle.mp3",
					Title: "Triangle",
					Size:  204299, Duration: 1999,
					Track: NewNullInt64(2), NumTracks: NewNullInt64(20),
					Disk: NewNullInt64(1), NumDisks: NewNullInt64(1),
					Artist: NewNullString("BADBADNOTGOOD")},
				Song{ID: 6, Album: NewNullInt64(1), Genre: NewNullInt64(1),
					Path:     "/home/test/Music/BADBADNOTGOOD/III/04 Something.mp3",
					Title:    "Something",
					Size:     91841,
					Duration: 9381,
					Artist:   NewNullString("BADBADNOTGOOD")}}},

		// order by single element
		{"lookup albums by album-artist", Album{Artist: NewNullInt64(1)}, []string{"num_tracks"},
			[]interface{}{
				Album{ID: 5, Artist: NewNullInt64(1), Year: NewNullInt64(2012),
					NumTracks: NewNullInt64(19), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1688), Title: "IV"},
				Album{ID: 1, Artist: NewNullInt64(1), Year: NewNullInt64(2011),
					NumTracks: NewNullInt64(20), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1688), Title: "III"},
			},
		},

		// order by multiple elemnts
		{"lookup albums by disk, order by duration, num_tracks",
			Album{NumDisks: NewNullInt64(1)}, []string{"duration", "num_tracks"},
			[]interface{}{
				Album{ID: 5, Artist: NewNullInt64(1), Year: NewNullInt64(2012),
					NumTracks: NewNullInt64(19), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1688), Title: "IV"},
				Album{ID: 1, Artist: NewNullInt64(1), Year: NewNullInt64(2011),
					NumTracks: NewNullInt64(20), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1688), Title: "III"},
				Album{ID: 4, Artist: NewNullInt64(4), Year: NewNullInt64(1985),
					NumTracks: NewNullInt64(13), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1756), Title: "Rust in Peace"},
				Album{ID: 2, Artist: NewNullInt64(2), Year: NewNullInt64(2001),
					NumTracks: NewNullInt64(10), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1800), Title: "Sour Soul"},
				Album{ID: 3, Artist: NewNullInt64(3), Year: NewNullInt64(1980),
					NumTracks: NewNullInt64(8), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(15440), Title: "Killers"},
			}},

		{"lookup albums by number of disks, order by duration, release year",
			Album{NumDisks: NewNullInt64(1)}, []string{"duration", "release_year"},
			[]interface{}{
				Album{ID: 1, Artist: NewNullInt64(1), Year: NewNullInt64(2011),
					NumTracks: NewNullInt64(20), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1688), Title: "III"},
				Album{ID: 5, Artist: NewNullInt64(1), Year: NewNullInt64(2012),
					NumTracks: NewNullInt64(19), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1688), Title: "IV"},
				Album{ID: 4, Artist: NewNullInt64(4), Year: NewNullInt64(1985),
					NumTracks: NewNullInt64(13), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1756), Title: "Rust in Peace"},
				Album{ID: 2, Artist: NewNullInt64(2), Year: NewNullInt64(2001),
					NumTracks: NewNullInt64(10), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(1800), Title: "Sour Soul"},
				Album{ID: 3, Artist: NewNullInt64(3), Year: NewNullInt64(1980),
					NumTracks: NewNullInt64(8), NumDisks: NewNullInt64(1),
					Duration: NewNullFloat64(15440), Title: "Killers"},
			}},

		{"lookup songs by size and genre, order by id",
			Song{Size: 91841, Genre: NewNullInt64(1)}, []string{"id"},
			[]interface{}{
				Song{ID: 6, Album: NullInt64{sql.NullInt64{1, true}},
					Genre: NullInt64{sql.NullInt64{1, true}},
					Path:  "/home/test/Music/BADBADNOTGOOD/III/04 Something.mp3",
					Title: "Something", Size: 91841, Duration: 9381,
					Track:     NullInt64{sql.NullInt64{0, false}},
					NumTracks: NullInt64{sql.NullInt64{0, false}},
					Disk:      NullInt64{sql.NullInt64{0, false}},
					NumDisks:  NullInt64{sql.NullInt64{0, false}},
					Artist:    NullString{sql.NullString{"BADBADNOTGOOD", true}}},
			}},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			prepareDB()
			results, err := wdb.Read(test.query, test.orderby)

			if err != nil {
				t.Error(err)
			}

			if len(results) != len(test.answer) {
				t.Fatalf("number of results do not match expected.\n\tresults: %#v\n\texpected: %v\n", results, test.answer)
			}

			for i := range results {
				if test.answer[i] != results[i] {
					t.Error(fmt.Errorf("test %s failed\n\t%9s %+v\n\t%-9s %+v",
						test.name, "expected:", test.answer[i], "result:", results[i]))
				}
			}
		})
	}
}

// TestAddItem ...
func TestAddItem(t *testing.T) {
	prepareDB()

	var (
		thirdCaseArtist = &Song{Artist: NewNullString("BADBADNOTGOOD")}
		thirdCaseErr    = ErrNonUnique{thirdCaseArtist}
	)

	testCases := [...]struct {
		query     interface{}
		returning []string
		answer    interface{}
		expErr    error
	}{
		// no genre
		{ // 0
			&Song{
				Album: NewNullInt64(1), Path: "/home/test/Music/BADBADNOTGOOD/III/03 Sax Stuff.mp3",
				Title: "Sax Stuff", Track: NewNullInt64(3), NumTracks: NewNullInt64(20),
				Disk: NewNullInt64(1), NumDisks: NewNullInt64(1), Size: 21134,
				Duration: 168, Artist: NewNullString("BADBADNOTGOOD & LeLand WILLY")},
			[]string{"id"},
			&Song{ID: 10001, Album: NewNullInt64(1), Genre: NullInt64{},
				Path:  "/home/test/Music/BADBADNOTGOOD/III/03 Sax Stuff.mp3",
				Title: "Sax Stuff", Track: NewNullInt64(3), NumTracks: NewNullInt64(20),
				Disk: NewNullInt64(1), NumDisks: NewNullInt64(1),
				Size: 21134, Duration: 168, Artist: NewNullString("BADBADNOTGOOD & LeLand WILLY")},
			nil,
		},
		// add a genre, then use it
		{ // 1
			&Genre{Name: "Jazz Hop"}, []string{"id"}, &Genre{10001, "Jazz Hop"}, nil,
		},
		{ // 2
			&Song{Album: NullInt64{}, Genre: NewNullInt64(10001),
				Path: "/home/test/Music/BADBADNOTGOOD/III/05 TT.mp3", Title: "TT",
				Track: NewNullInt64(5), NumTracks: NewNullInt64(20),
				Disk: NewNullInt64(1), NumDisks: NewNullInt64(1), Size: 2841,
				Duration: 111, Artist: NullString{}},
			[]string{"id"},
			&Song{ID: 10002, Album: NullInt64{}, Genre: NewNullInt64(10001),
				Path: "/home/test/Music/BADBADNOTGOOD/III/05 TT.mp3", Title: "TT",
				Track: NewNullInt64(5), NumTracks: NewNullInt64(20),
				Disk: NewNullInt64(1), NumDisks: NewNullInt64(1), Size: 2841,
				Duration: 111, Artist: NullString{}},
			nil,
		},

		// don't provide path for lookup
		{ // 3
			thirdCaseArtist,
			[]string{"id"},
			&Song{Artist: NewNullString("BADBADNOTGOOD")},
			thirdCaseErr,
		},

		{ // 4
			&Song{Path: "/", Title: "test", Size: 444, Duration: 4445,
				Artist: NewNullString("BADBADNOTGOOD")},
			[]string{"id"},
			&Song{ID: 10003, Path: "/", Title: "test", Size: 444, Duration: 4445,
				Artist: NewNullString("BADBADNOTGOOD")},
			nil,
		},

		{ // 5
			&Song{
				Album: NewNullInt64(1), Path: "/home/test/Music/BADBADNOTGOOD/III/03 Sax Stuff.mp3",
				Title: "Sax Stuff", Track: NewNullInt64(3), NumTracks: NewNullInt64(20),
				Disk: NewNullInt64(1), NumDisks: NewNullInt64(1), Size: 21134,
				Duration: 168, Artist: NewNullString("BADBADNOTGOOD & LeLand WILLY")},
			[]string{"id"},
			&Song{ID: 10001, Album: NewNullInt64(1), Genre: NullInt64{},
				Path:  "/home/test/Music/BADBADNOTGOOD/III/03 Sax Stuff.mp3",
				Title: "Sax Stuff", Track: NewNullInt64(3), NumTracks: NewNullInt64(20),
				Disk: NewNullInt64(1), NumDisks: NewNullInt64(1),
				Size: 21134, Duration: 168, Artist: NewNullString("BADBADNOTGOOD & LeLand WILLY")},
			ErrAlreadyExists,
		},
	}

	for testCase, test := range testCases {
		err := wdb.Create(test.query, test.returning)

		if err != test.expErr {
			t.Errorf("test case %d failed: %v\n", testCase, err)
		}

		// reflection is the only way to compare these as they are pointers
		rQ := reflect.ValueOf(test.query)
		rA := reflect.ValueOf(test.answer)
		if rQ.Elem().Interface() != rA.Elem().Interface() {
			t.Errorf("test case %d failed:\n\texpected: %v\n\tresult:   %v\n", testCase, test.answer, test.query)
		}
	}
}

// TestUpdate ...
func TestUpdate(t *testing.T) {
	testCases := [...]struct {
		name         string
		set          interface{}
		where        interface{}
		expErr       error
		answerLookup interface{}
		answer       []interface{}
	}{
		{"update first song", Song{Title: "My Knight"}, Song{ID: 1}, nil,
			Song{ID: 1},
			[]interface{}{
				Song{ID: 1, Album: NullInt64{sql.NullInt64{1, true}},
					Genre: NullInt64{sql.NullInt64{1, true}},
					Path:  "/home/test/Music/BADBADNOTGOOD/III/01 In the Night.mp3",
					Title: "My Knight", Size: 204192, Duration: 1993,
					Track:     NullInt64{sql.NullInt64{1, true}},
					NumTracks: NullInt64{sql.NullInt64{20, true}},
					Disk:      NullInt64{sql.NullInt64{1, true}},
					NumDisks:  NullInt64{sql.NullInt64{1, true}},
					Artist:    NullString{sql.NullString{"BADBADNOTGOOD", true}}},
			}},

		{"update multiple fields", Song{NumTracks: NewNullInt64(20), Track: NewNullInt64(4)},
			Song{Size: 91841, Genre: NewNullInt64(1)}, nil,
			Song{Size: 91841, Genre: NewNullInt64(1)},
			[]interface{}{
				Song{ID: 6, Album: NullInt64{sql.NullInt64{1, true}},
					Genre: NullInt64{sql.NullInt64{1, true}},
					Path:  "/home/test/Music/BADBADNOTGOOD/III/04 Something.mp3",
					Title: "Something", Size: 91841, Duration: 9381,
					Track:     NullInt64{sql.NullInt64{4, true}},
					NumTracks: NullInt64{sql.NullInt64{20, true}},
					Disk:      NullInt64{sql.NullInt64{0, false}},
					NumDisks:  NullInt64{sql.NullInt64{0, false}},
					Artist:    NullString{sql.NullString{"BADBADNOTGOOD", true}}},
			}},

		{"update multiple songs' artist using artist as query", Song{Artist: NewNullString("BED BED NUT GUD")},
			Song{Artist: NewNullString("BADBADNOTGOOD")}, nil,
			Song{Artist: NewNullString("BED BED NUT GUD")},
			[]interface{}{
				Song{ID: 1, Album: NullInt64{sql.NullInt64{1, true}},
					Genre: NullInt64{sql.NullInt64{1, true}},
					Path:  "/home/test/Music/BADBADNOTGOOD/III/01 In the Night.mp3",
					Title: "In the Night", Size: 204192, Duration: 1993,
					Track:     NullInt64{sql.NullInt64{1, true}},
					NumTracks: NullInt64{sql.NullInt64{20, true}},
					Disk:      NullInt64{sql.NullInt64{1, true}},
					NumDisks:  NullInt64{sql.NullInt64{1, true}},
					Artist:    NullString{sql.NullString{"BED BED NUT GUD", true}}},
				Song{ID: 5, Album: NullInt64{sql.NullInt64{1, true}},
					Genre: NullInt64{sql.NullInt64{1, true}},
					Path:  "/home/test/Music/BADBADNOTGOOD/III/02 Triangle.mp3",
					Title: "Triangle", Size: 204299, Duration: 1999,
					Track:     NullInt64{sql.NullInt64{2, true}},
					NumTracks: NullInt64{sql.NullInt64{20, true}},
					Disk:      NullInt64{sql.NullInt64{1, true}},
					NumDisks:  NullInt64{sql.NullInt64{1, true}},
					Artist:    NullString{sql.NullString{"BED BED NUT GUD", true}}},
				Song{ID: 6, Album: NullInt64{sql.NullInt64{1, true}},
					Genre: NullInt64{sql.NullInt64{1, true}},
					Path:  "/home/test/Music/BADBADNOTGOOD/III/04 Something.mp3",
					Title: "Something", Size: 91841, Duration: 9381,
					Track:     NullInt64{sql.NullInt64{0, false}},
					NumTracks: NullInt64{sql.NullInt64{0, false}},
					Disk:      NullInt64{sql.NullInt64{0, false}},
					NumDisks:  NullInt64{sql.NullInt64{0, false}},
					Artist:    NullString{sql.NullString{"BED BED NUT GUD", true}}}},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			prepareDB()

			err := wdb.Update(test.set, test.where)
			if test.expErr != err {
				t.Errorf("expected error did not match received\n\texpected: %v\n\treceived: %v",
					test.expErr, err)
			}

			result, err := wdb.Read(test.answerLookup, []string{})
			if len(result) != len(test.answer) {
				t.Fatalf("expected length of results did not match length of answer\n\tresults: %v\n\tanswer: %v\n",
					result, test.answer)
			}

			for i := range result {
				if result[i] != test.answer[i] {
					t.Errorf("result %d did not match answer\n\tresult: %v\n\tanswer: %v\n",
						i, result[i], test.answer[i])
				}
			}
		})
	}
}
