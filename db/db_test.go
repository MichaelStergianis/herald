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

	"github.com/lib/pq"
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

	// songs, err := hdb.GetSongsInLibrary(libs[testLibName])
	// if err != nil {
	// 	t.Error(err)
	// }

	// fmt.Printf("%+v\n", songs)

	// for _, song := range songs {
	// 	fmt.Printf("%+v\n", song)
	// }
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
		{Song{Artist: NewNullString("BADBADNOTGOOD")}, "edn", []string{},
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
		{Album{Artist: NewNullInt64(1)}, "json", []string{"num-tracks"},
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
		{Album{NumDisks: NewNullInt64(1)}, "edn", []string{"duration", "num-tracks"},
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
		{Album{NumDisks: NewNullInt64(1)}, "edn", []string{"duration", "year"},
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
		{
			&Genre{Name: "Jazz Hop"}, []string{"id"}, &Genre{10001, "Jazz Hop"}, nil,
		},
		{
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
		{
			&Song{Artist: NewNullString("BADBADNOTGOOD")},
			[]string{"id"},
			&Song{Artist: NewNullString("BADBADNOTGOOD")},
			&pq.Error{Severity: "ERROR", Code: "23502",
				Message: "null value in column \"fs_path\" violates not-null constraint",
				Detail: "Failing row contains (10003, null, null, null, " +
					"null, null, null, null, null, null, null, BADBADNOTGOOD).",
				Hint: "", Position: "", InternalPosition: "", InternalQuery: "",
				Where: "", Schema: "music", Table: "songs", Column: "fs_path",
				DataTypeName: "", Constraint: "", File: "execMain.c", Line: "2042", Routine: "ExecConstraints"},
		},

		{
			&Song{Path: "/", Title: "test", Size: 444, Duration: 4445,
				Artist: NewNullString("BADBADNOTGOOD")},
			[]string{"id"},
			&Song{ID: 10004, Path: "/", Title: "test", Size: 444, Duration: 4445,
				Artist: NewNullString("BADBADNOTGOOD")},
			nil,
		},
	}

	for testCase, test := range testCases {
		q, err := hdb.addItem(test.query, test.returning)
		fmt.Printf("testCase = %v\n", testCase)

		fmt.Printf("%T %v\n", q, q)
		fmt.Printf("%T %v\n\n", test.query, test.query)

		// test case 4 (index 3) is a special case, returns a pointer to an
		// error struct, so equality testing is difficult
		var secondaryErrCheck = true
		if _, ok := test.expErr.(*pq.Error); ok {
			secondaryErrCheck = false
			if *err.(*pq.Error) != *test.expErr.(*pq.Error) {
				t.Errorf("test case %d failed: %v\n", testCase, err)
			}
		}

		if err != test.expErr && secondaryErrCheck {
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

// TestExtra ...
func TestExtra(t *testing.T) {
	prepareDB()

	for _, col := range []interface{}{&Album{}, &Artist{}, &Song{}, &Library{}} {
		fmt.Printf("\n%v\n", reflect.TypeOf(col))

		cols, err := hdb.GetItem(col, []string{"id"})

		if err != nil {
			t.Error(err)
		}

		for i, c := range cols {
			fmt.Printf("\t%d %#v\n", i, c)

		}
	}
}
