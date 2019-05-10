package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var (
	hdb       *HeraldDB
	prepareDB func()
)

const (
	testLibName = "Test"
	dbName      = "herald_test"
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

// TestGetArtists ...
func TestGetArtists(t *testing.T) {
	prepareDB()

	expected := []Artist{
		Artist{ID: 1, Name: "BADBADNOTGOOD", Path: "/home/test/Music/BADBADNOTGOOD"},
		Artist{ID: 2, Name: "BADBADNOTGOOD & Ghostface Killah", Path: "/home/test/Music/BADBADNOTGOOD & Ghostface Killah"},
		Artist{ID: 3, Name: "Iron Maiden", Path: "/home/tests/MyMusic/Iron Maiden"},
		Artist{ID: 4, Name: "Megadeth", Path: "/home/test/Music/Megadeth"},
	}

	results, err := hdb.GetArtists()
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(results); i++ {
		if expected[i] != results[i] {
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

// TestGetUniqueAlbum ...
func TestGetUniqueAlbum(t *testing.T) {
	prepareDB()

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
		Path:      "/home/test/Music/BADBADNOTGOOD/III",
		Duration:  1688,
	}

	a, err = hdb.GetUniqueAlbum(album)

	if err != nil {
		t.Error(err)
	}

	if a != album {
		t.Error(errors.New("expected album is not in database"))
	}
}

// TestGetUniqueAlbumNegative ...
func TestGetUniqueAlbumNegative(t *testing.T) {
	prepareDB()

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

	a, err = hdb.GetUniqueAlbum(album)

	if err != ErrNotPresent && err != nil {
		t.Error(err)
	}

	if a == album {
		t.Error("unexpected album is in database")
	}
}

// TestAddArtist ...
func TestAddArtist(t *testing.T) {
	prepareDB()

	artist := Artist{
		Name: "Clever Girl",
		Path: path.Join(testLib, "Clever Girl/"),
	}

	a, err := hdb.addArtist(artist)

	if err != nil {
		t.Error(err)
	}

	a.ID = 0

	if a != artist {
		t.Error("unexpected value returned")

	}
}

// TestAddAlbum ...
func TestAddAlbum(t *testing.T) {
	prepareDB()

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
		t.Error("unexpected value returned")
	}
}

// TestDuration ...
func TestDuration(t *testing.T) {
	prepareDB()
	prepareTestLibrary()

	song := Song{
		Path: path.Join(testLib, "TTNG/Animals Acoustic/05 Quetzal.mp3"),
	}

	d, err := duration(song)

	if err != nil {
		t.Error(err)
	}

	if d != 87.22 {
		t.Error("unexpected return value")
	}
}

// TestAddSong ...
func TestAddSong(t *testing.T) {
	prepareDB()

	song := Song{
		Album:     1,
		Genre:     1,
		Path:      "/home/test/Music/BADBADNOTGOOD/III/02 Something.mp3",
		Title:     "Something",
		Track:     2,
		NumTracks: 20,
		Disk:      1,
		NumDisks:  1,
		Size:      999,
		Duration:  998,
		Artist:    "BADBADNOTGOOD",
	}

	s, err := hdb.addSong(song)

	if err != nil {
		t.Error(err)
	}

	songVerify, err := hdb.GetUniqueSong(s)

	if err != nil {
		t.Error(err)
	}

	if s != songVerify {
		t.Error("song was not added to library correctly")
	}
}

// TestGetSongsInLibrary ...
func TestGetSongsInLibrary(t *testing.T) {
	prepareDB()

	songsGroundTruth := []Song{
		Song{ID: 1, Album: 1, Genre: 1,
			Path:  "/home/test/Music/BADBADNOTGOOD/III/01 In the Night.mp3",
			Title: "In the Night",
			Track: 1, NumTracks: 20, Disk: 1, NumDisks: 1,
			Size: 204192, Duration: 1993,
			Artist: "BADBADNOTGOOD"},
		Song{ID: 2, Album: 2, Genre: 2,
			Path:  "/home/test/Music/BADBADNOTGOOD & Ghostface Killah/Sour Soul/01 Sour Soul.mp3",
			Title: "Sour Soul",
			Track: 1, NumTracks: 10, Disk: 1, NumDisks: 1,
			Size: 19203, Duration: 1920,
			Artist: "BADBADNOTGOOD & Ghostface Killah"},
		Song{ID: 4, Album: 4, Genre: 4,
			Path:  "/home/test/Music/Megadeth/Rust in Peace/01 Hangar 18.mp3",
			Title: "Hangar 18",
			Track: 1, NumTracks: 13, Disk: 1, NumDisks: 1,
			Size: 99948, Duration: 9994,
			Artist: "Megadeth"},
	}

	libs, err := hdb.GetLibraries()
	if err != nil {
		t.Error(err)
	}

	songs, err := hdb.GetSongsInLibrary(libs["Music"])
	if err != nil {
		t.Error(err)
	}

	if len(songs) != len(songsGroundTruth) {
		t.Error("too many songs")
	}

	for idx, song := range songs {
		if song != songsGroundTruth[idx] {
			t.Error("unexpected song")
		}
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

	testSong := path.Join(testLib, "GoldLink/At What Cost/02 Same Clothes as Yesterday.m4a")
	err = hdb.processMedia(testSong, libs[testLibName])
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

	fmt.Printf("%v\n", libs)

	err = hdb.ScanLibrary(libs[testLibName])
	if err != nil {
		t.Error(err)
	}

	// songs, err := hdb.GetSongsInLibrary(libs[testLibName])
	// if err != nil {
	// 	t.Error(err)
	// }

	// fmt.Printf("%v\n", songs)

	// for _, song := range songs {
	// 	fmt.Printf("%+v\n", song)
	// }
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
		results, err := hdb.GetItem(test.query, converter, test.orderby)

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
