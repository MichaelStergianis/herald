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

	// pq is used behind the scenes, but never explicitly used
	_ "github.com/lib/pq"

	"github.com/dhowden/tag"
	ft "gopkg.in/h2non/filetype.v1"
)

// HeraldDB ...
// A type for interfacing with the herald db
type HeraldDB struct {
	*sql.DB
}

// check ...
func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}

// Open ...
// Creates the connection to the db as a HeraldDB pointer.
func Open(connStr string) (*HeraldDB, error) {
	sqldb, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	hdb := &HeraldDB{
		sqldb,
	}
	return hdb, nil
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
		return d, errors.New("hdb: no duration for empty file")
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
// Closes an hdb.
func (hdb *HeraldDB) Close() {
	hdb.DB.Close()
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
func (hdb *HeraldDB) CountTable(table string) (count int, err error) {
	if ok := GetValidTable(table); !ok {
		return 0, ErrInvalidTable
	}

	row := hdb.QueryRow(`SELECT COUNT(1) AS count FROM ` + table)

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
func (hdb *HeraldDB) AddLibrary(name string, fsPath string) (err error) {
	if !path.IsAbs(fsPath) {
		return ErrNotAbs
	}

	stmt, err := hdb.Prepare("INSERT INTO music.libraries (name, fs_path) VALUES ($1, $2);")
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
func (hdb *HeraldDB) GetLibraries() (libs map[string]Library, err error) {
	tableName := "music.libraries"

	count, err := hdb.CountTable(tableName)

	// query
	rows, err := hdb.Query("SELECT id, name, fs_path from " + tableName + " ORDER BY id;")
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
func (hdb *HeraldDB) processMedia(fsPath string, lib Library) (err error) {
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
	inLib, err := hdb.songInLibrary(*s, lib)
	if err != nil {
		return err
	}

	if inLib {
		return nil
	}

	s.Track, s.NumTracks = metadata.Track()
	s.Disk, s.NumDisks = metadata.Disc()

	s.Duration, err = duration(*s)
	if err != nil {
		return err
	}

	// add genre information
	genre := &Genre{
		Name: metadata.Genre(),
	}
	if genre.Name != "" {
		_, err = hdb.addItem(genre, []string{"id"})
		if err != nil {
			return err
		}
	}

	// Add the album artist information
	artist := &Artist{
		Name: metadata.AlbumArtist(),
		Path: stripToArtist(fsPath, lib),
	}
	if artist.Name != "" {
		_, err = hdb.addItem(artist, []string{"id"})
		if err != nil {
			return err
		}
	}

	// Add the album information
	album := &Album{
		Path:      stripToAlbum(fsPath, *artist),
		Artist:    artist.ID,
		Year:      metadata.Year(),
		NumTracks: s.NumTracks,
		NumDisks:  s.NumDisks,
		Title:     metadata.Album(),
	}
	_, err = hdb.addItem(album, []string{"id"})
	if err != nil {
		return err
	}

	s.Album = album.ID
	s.Genre = genre.ID

	_, err = hdb.addItem(s, []string{"id"})
	if err != nil {
		return err
	}

	err = hdb.addSongToLibrary(*s, lib)
	if err != nil {
		return err
	}

	return nil
}

// songInLibrary ...
// Checks to see if a song is in the given library.
func (hdb *HeraldDB) songInLibrary(song Song, library Library) (inLib bool, err error) {
	// get songs id based on path
	if song.Path == "" {
		return false, ErrNonUnique{song}
	}
	err = hdb.QueryRow("SELECT id FROM music.songs where fs_path = $1", song.Path).Scan(&song.ID)
	if err != nil && !(err == sql.ErrNoRows) {
		return false, err
	}

	query := "SELECT COUNT(1) FROM music.songs_in_library where song_id = $1 AND library_id = $2;"

	row := hdb.QueryRow(query, song.ID, library.ID)

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
func (hdb *HeraldDB) GetSongsInLibrary(lib Library) (songs []Song, err error) {
	if lib.ID == 0 {
		return nil, errors.New("hdb: provided library must have an id")
	}

	var size int
	err = hdb.QueryRow("SELECT COUNT(1) FROM music.songs_in_library WHERE library_id = $1", lib.ID).Scan(&size)
	if err != nil {
		return nil, err
	}
	songs = make([]Song, size)

	rows, err := hdb.Query("SELECT song_id FROM music.songs_in_library WHERE library_id = $1 ORDER BY song_id", lib.ID)
	if err != nil {
		return nil, err
	}
	for idx := 0; rows.Next(); idx++ {
		var s Song
		err = rows.Scan(&s.ID)
		if err != nil {
			return nil, err
		}

		err := hdb.GetUniqueItem(&s)
		if err != nil {
			return nil, err
		}
		songs[idx] = s
	}

	return songs, nil
}

// addSongToLibrary ...
func (hdb *HeraldDB) addSongToLibrary(song Song, lib Library) error {
	sInL, err := hdb.songInLibrary(song, lib)
	if err != nil {
		return err
	}

	if sInL {
		return nil
	}

	query := "INSERT INTO music.songs_in_library (song_id, library_id) VALUES ($1, $2)"
	stmt, err := hdb.Prepare(query)
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
func (hdb *HeraldDB) addImageFile(fsPath string) {

}

// ScanLibrary ...
// Scans the library. If some media is already in the library, it will not add it again.
func (hdb *HeraldDB) ScanLibrary(lib Library) (err error) {
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
				// fmt.Printf("lib: %v, song: %v\n", lib, fsPath)

				err = hdb.processMedia(fsPath, lib)
				if err != nil {
					log.Printf("%v", err)
				}
			}
		case imageType:
			hdb.addImageFile(fsPath)
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
func (hdb *HeraldDB) ScanLibraries() {
	libs, err := hdb.GetLibraries()

	check(err)

	for _, lib := range libs {
		hdb.ScanLibrary(lib)
	}
}
