package db

import (
	"bytes"
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/dhowden/tag"
	// pq is used behind the scenes, but never explicitly used
	_ "github.com/lib/pq"

	ft "gopkg.in/h2non/filetype.v1"
)

// WarblerDB ...
// A type for interfacing with the warbler db
type WarblerDB struct {
	*sql.DB
}

// check ...
func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}

// Open ...
// Creates the connection to the db as a WarblerDB pointer.
func Open(connStr string) (*WarblerDB, error) {
	sqldb, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	wdb := &WarblerDB{
		sqldb,
	}
	return wdb, nil
}

// duration ...
// Uses ffmpeg to get songs duration.
func duration(song Song) (d float64, err error) {
	cmd := exec.Command("ffprobe", song.Path)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return d, err
	}

	// parse the stderr of ffprobe
	query := "Duration: ([0-9]{2}):([0-9]{2}):([0-9]{2})[.]([0-9]{1,})"
	re, err := regexp.Compile(query)
	if err != nil {
		return d, err
	}

	matches := re.FindAllStringSubmatch(string(stderr.Bytes()), -1)
	if matches == nil {
		return d, errors.New("wdb: no duration for empty file")
	}

	timings := make([]float64, 4)

	for i := 0; i < 4; i++ {
		timings[i], err = strconv.ParseFloat(matches[0][i+1], 64)
		if err != nil {
			return d, err
		}
	}

	decimals := math.Pow(10, -math.Ceil(math.Log10(timings[3])))

	d = timings[0]*3600 + timings[1]*60 + timings[2] + timings[3]*decimals

	return d, nil
}

// Close ...
// Closes an wdb.
func (wdb *WarblerDB) Close() {
	wdb.DB.Close()
}

// GetValidTable ...
// Checks to see if the table passed to CountTable is in the list of valid tables.
func GetValidTable(table string) bool {
	// create an empty type for our set
	type empty struct{}
	var validTables = map[string]struct{}{
		// music schema
		"music.artists":   empty{},
		"music.genres":    empty{},
		"music.images":    empty{},
		"music.albums":    empty{},
		"music.songs":     empty{},
		"music.libraries": empty{},

		// config schema
		// "config.preferences": empty{},
		// "config.users":       empty{},

		// Multiple IDs
		"music.images_in_album":  empty{},
		"music.songs_in_library": empty{},
	}

	_, ok := validTables[table]
	return ok
}

// CountTable ...
// Gets the count of a table in our database.
func (wdb *WarblerDB) CountTable(table string) (count int, err error) {
	if ok := GetValidTable(table); !ok {
		return 0, ErrInvalidTable
	}

	row := wdb.QueryRow(`SELECT COUNT(1) AS count FROM ` + table)

	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// AddLibrary ...
// Creates the library of a given name and path. Requries an absolute
// path. You should not make assumptions about from which directory
// this server will be run.
func (wdb *WarblerDB) AddLibrary(name string, fsPath string) (err error) {
	if !path.IsAbs(fsPath) {
		return ErrNotAbs
	}

	stmt, err := wdb.Prepare("INSERT INTO music.libraries (name, fs_path) VALUES ($1, $2);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, fsPath)
	if err != nil {
		return err
	}

	return nil
}

// GetLibraries ...
func (wdb *WarblerDB) GetLibraries() (libs map[string]Library, err error) {
	tableName := "music.libraries"

	count, err := wdb.CountTable(tableName)

	// query
	rows, err := wdb.Query("SELECT id, name, fs_path from " + tableName + " ORDER BY id;")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	libs = make(map[string]Library, count)
	for i := 0; rows.Next(); i++ {
		var l Library
		err = rows.Scan(&l.ID, &l.Name, &l.Path)
		if err != nil {
			return nil, err
		}

		libs[l.Name] = l
	}

	return libs, nil
}

// checkFileType ...
func fileType(file string) int {
	buf, err := ioutil.ReadFile(file)
	check(err)

	if ft.IsAudio(buf) {
		return musicType
	} else if ft.IsImage(buf) {
		return imageType
	} else {
		return unknownType
	}
}

// processMedia ...
// Processes a file to add necessary information to the database.
func (wdb *WarblerDB) processMedia(fsPath string, lib Library) (err error) {
	f, err := os.Open(fsPath)
	if err != nil {
		return err
	}

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		return err
	}

	stats, err := f.Stat()
	if err != nil {
		return err
	}

	s := &Song{
		Path:  fsPath,
		Title: metadata.Title(),
		Size:  stats.Size(),
	}

	// Check to see if the song is in the database
	inLib, err := wdb.songInLibrary(*s, lib)
	if err != nil {
		return err
	}

	if inLib {
		return nil
	}

	t, nT := metadata.Track()
	d, nD := metadata.Disc()
	sqlInts := []*NullInt64{&s.Track, &s.NumTracks, &s.Disk, &s.NumDisks}
	for i, v := range []int{t, nT, d, nD} {
		sqlInts[i].Int64 = int64(v)
		if sqlInts[i].Int64 != 0 {
			sqlInts[i].Valid = true
		}
	}

	s.Duration, err = duration(*s)
	if err != nil {
		return err
	}

	// add genre information
	genre := &Genre{
		Name: metadata.Genre(),
	}
	if genre.Name != "" {
		err = wdb.Create(genre, []string{"id"})
		if err != nil && err != ErrAlreadyExists {
			return err
		}
	}

	// Add the album artist information
	artist := &Artist{
		Name: metadata.AlbumArtist(),
	}
	if artist.Name != "" {
		err = wdb.Create(artist, []string{"id"})
		if err != nil && err != ErrAlreadyExists {
			return err
		}
	}

	albumYear := NewNullInt64(int64(metadata.Year()))
	var albumArtist NullInt64
	if artist.ID != 0 {
		albumArtist.Int64 = artist.ID
		albumArtist.Valid = true
	}

	// Add the album information
	album := &Album{
		Artist:    albumArtist,
		Year:      albumYear,
		NumTracks: s.NumTracks,
		NumDisks:  s.NumDisks,
		Title:     metadata.Album(),
	}

	err = wdb.Create(album, []string{"id"})
	if err != nil && err != ErrAlreadyExists {
		return err
	}

	if album.ID != 0 {
		s.Album = NewNullInt64(album.ID)
	}
	if genre.ID != 0 {
		s.Genre = NewNullInt64(genre.ID)
	}
	for _, v := range []*NullInt64{&s.Album, &s.Genre} {
		if v.Int64 != 0 {
			v.Valid = true
		}
	}

	err = wdb.Create(s, []string{"id"})
	if err != nil && err != ErrAlreadyExists {
		return err
	}

	err = wdb.addSongToLibrary(*s, lib)
	if err != nil {
		return err
	}

	return nil
}

// songInLibrary ...
// Checks to see if a song is in the given library.
func (wdb *WarblerDB) songInLibrary(song Song, library Library) (inLib bool, err error) {
	// get songs id based on path
	if song.Path == "" {
		return false, ErrNonUnique{song}
	}
	err = wdb.QueryRow("SELECT id FROM music.songs where fs_path = $1", song.Path).Scan(&song.ID)
	if err != nil && !(err == sql.ErrNoRows) {
		return false, err
	}

	query := "SELECT COUNT(1) FROM music.songs_in_library where song_id = $1 AND library_id = $2;"

	row := wdb.QueryRow(query, song.ID, library.ID)

	var numInLib int
	err = row.Scan(&numInLib)

	if err != nil {
		return false, err
	}

	if numInLib > 1 {
		return false, ErrNonUnique{library}
	}

	// convert 1/0 to true/false
	inLib = numInLib == 1

	return inLib, nil
}

// TODO GetTypeInLibrary ...

// GetSongsInLibrary ...
func (wdb *WarblerDB) GetSongsInLibrary(lib Library) (songs []Song, err error) {
	if lib.ID == 0 {
		return nil, errors.New("wdb: provided library must have an id")
	}

	var size int
	err = wdb.QueryRow("SELECT COUNT(1) FROM music.songs_in_library WHERE library_id = $1", lib.ID).Scan(&size)
	if err != nil {
		return nil, err
	}
	songs = make([]Song, size)

	rows, err := wdb.Query("SELECT song_id FROM music.songs_in_library WHERE library_id = $1 ORDER BY song_id", lib.ID)
	if err != nil {
		return nil, err
	}
	for idx := 0; rows.Next(); idx++ {
		var s Song
		err = rows.Scan(&s.ID)
		if err != nil {
			return nil, err
		}

		err := wdb.ReadUnique(&s)
		if err != nil {
			return nil, err
		}
		songs[idx] = s
	}

	return songs, nil
}

// addSongToLibrary ...
func (wdb *WarblerDB) addSongToLibrary(song Song, lib Library) error {
	sInL, err := wdb.songInLibrary(song, lib)
	if err != nil {
		return err
	}

	if sInL {
		return nil
	}

	query := "INSERT INTO music.songs_in_library (song_id, library_id) VALUES ($1, $2)"
	stmt, err := wdb.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(&song.ID, &lib.ID)
	if err != nil {
		return err
	}

	return nil
}

// addImageFile ...
func (wdb *WarblerDB) addImageFile(fsPath string) {

}

// ScanLibrary ...
// Scans the library. If some media is already in the library, it will not add it again.
func (wdb *WarblerDB) ScanLibrary(lib Library) (err error) {
	walkFn := func(fsPath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Encountered the following error while traversing %q: %v", fsPath, err)
			return err
		}
		if info.IsDir() {
			return err
		}
		switch fileType(fsPath) {
		case musicType:
			{
				err = wdb.processMedia(fsPath, lib)
				if err != nil {
					log.Printf("%v", err)
				}
			}
		case imageType:
			wdb.addImageFile(fsPath)
		}
		return err
	}

	err = filepath.Walk(lib.Path, walkFn)
	if err != nil {
		return err
	}

	return nil
}

// ScanLibraries ...
// Scans all available libraries
func (wdb *WarblerDB) ScanLibraries() {
	libs, err := wdb.GetLibraries()

	check(err)

	for _, lib := range libs {
		wdb.ScanLibrary(lib)
	}
}
