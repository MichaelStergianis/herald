package db

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	// pq is used behind the scenes, but never explicitly used
	_ "github.com/lib/pq"

	"github.com/dhowden/tag"
	ft "gopkg.in/h2non/filetype.v1"
)

// HeraldDB ...
// A type for interfacing with the herald db
type HeraldDB struct {
	db *sql.DB
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

	hdb := HeraldDB{
		sqldb,
	}
	return &hdb, nil
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
	hdb.db.Close()
}

// isValidTable ...
// Checks to see if the table passed to CountTable is in the list of valid tables.
func isValidTable(table string) bool {
	// create an empty type for our set
	type empty struct{}

	var validTables = map[string]empty{
		// music schema
		"music.artists":          empty{},
		"music.genres":           empty{},
		"music.images":           empty{},
		"music.albums":           empty{},
		"music.images_in_album":  empty{},
		"music.songs":            empty{},
		"music.libraries":        empty{},
		"music.songs_in_library": empty{},

		// config schema
		"config.preferences": empty{},
		"config.users":       empty{},
	}

	_, inTable := validTables[table]
	return inTable
}

// CountTable ...
// Gets the count of a table in our database.
func (hdb *HeraldDB) CountTable(table string) (count int, err error) {
	if !isValidTable(table) {
		return 0, ErrInvalidTable
	}

	row := hdb.db.QueryRow(`SELECT COUNT(1) AS count FROM ` + table)

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

	stmt, err := hdb.db.Prepare("INSERT INTO music.libraries (name, fs_path) VALUES ($1, $2);")
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
	rows, err := hdb.db.Query("SELECT id, name, fs_path from " + tableName + " ORDER BY id;")
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

// select artists.name, albums.title from music.albums inner join
// music.artists on (music.albums.artist = music.artists.id);

// GetArtists ...
// Gets all artists in the database
func (hdb *HeraldDB) GetArtists() (artists []Artist, err error) {
	tableName := "music.artists"

	count, err := hdb.CountTable(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := hdb.db.Query("SELECT id, name, fs_path FROM " + tableName + " ORDER BY id;")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artists = make([]Artist, count)
	for i := 0; rows.Next(); i++ {
		var a Artist
		err = rows.Scan(&a.ID, &a.Name, &a.Path)

		if err != nil {
			return nil, err
		}

		artists[i] = a
	}

	return artists, nil
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

	var s Song
	s = Song{
		Path:  fsPath,
		Title: metadata.Title(),
	}

	// Check to see if the song is in the database
	inLib, err := hdb.songInLibrary(s, lib)
	if err != nil {
		return err
	}

	if inLib {
		return nil
	}

	s.Track, s.NumTracks = metadata.Track()
	s.Disk, s.NumDisks = metadata.Disc()

	s.Duration, err = duration(s)
	if err != nil {
		return err
	}

	// add genre information
	genre := Genre{
		Name: metadata.Genre(),
	}
	genre, err = hdb.addGenre(genre)
	if err != nil {
		return err
	}

	// Add the album artist information
	artist, err := hdb.addArtist(Artist{
		Name: metadata.AlbumArtist(),
		Path: stripToArtist(fsPath, lib),
	})
	if err != nil {
		return err
	}

	// Add the album information
	album, err := hdb.addAlbum(Album{
		Path:      stripToAlbum(fsPath, artist),
		Artist:    artist.ID,
		Year:      metadata.Year(),
		NumTracks: s.NumTracks,
		NumDisks:  s.NumDisks,
		Title:     metadata.Album(),
	})
	if err != nil {
		return err
	}

	s.Album = album.ID

	_, err = hdb.addSong(s)
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
		return false, ErrNonUnique
	}
	err = hdb.db.QueryRow("SELECT id FROM music.songs where fs_path = $1", song.Path).Scan(&song.ID)
	if err != nil && !(err == sql.ErrNoRows) {
		return false, err
	}

	query := "SELECT COUNT(1) FROM music.songs_in_library where song_id = $1 AND library_id = $2;"

	row := hdb.db.QueryRow(query, song.ID, library.ID)

	var numInLib int
	err = row.Scan(&numInLib)

	if err != nil {
		return false, err
	}

	if numInLib > 1 {
		return false, ErrNonUnique
	}

	// convert 1/0 to true/false
	inLib = numInLib == 1

	return inLib, nil
}

// GetSongsInLibrary ...
func (hdb *HeraldDB) GetSongsInLibrary(lib Library) (songs []Song, err error) {
	if lib.ID == 0 {
		return nil, errors.New("hdb: provided library must have an id")
	}

	var size int
	err = hdb.db.QueryRow("SELECT COUNT(1) FROM music.songs_in_library WHERE library_id = $1", lib.ID).Scan(&size)
	if err != nil {
		return nil, err
	}
	songs = make([]Song, size)

	rows, err := hdb.db.Query("SELECT song_id FROM music.songs_in_library WHERE library_id = $1 ORDER BY song_id", lib.ID)
	if err != nil {
		return nil, err
	}
	for idx := 0; rows.Next(); idx++ {
		var s Song
		err = rows.Scan(&s.ID)
		if err != nil {
			return nil, err
		}

		s, err := hdb.GetUniqueSong(s)
		if err != nil {
			return nil, err
		}
		songs[idx] = s
	}

	return songs, nil
}

// addSong ...
func (hdb *HeraldDB) addSong(song Song) (s Song, err error) {
	s, err = hdb.GetUniqueSong(song)
	if err != nil && err != ErrNotPresent {
		return Song{}, err
	}

	if s != (Song{}) {
		return s, nil
	}

	// add the song
	query := "INSERT INTO music.songs " +
		"(album, genre, fs_path, title, track, num_tracks, disk, num_disks, song_size, duration, artist) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"

	err = hdb.db.QueryRow(query, song.Album,
		song.Genre, song.Path, song.Title, song.Track, song.NumTracks, song.Disk,
		song.NumDisks, song.Size, song.Duration, song.Artist).Scan(&song.ID)

	if err != nil {
		return Song{}, err
	}

	return song, nil
}

// addAlbum ...
// looks in the database for the album information contained in the song metadata,
// if it is not found the function creates and returns the album
func (hdb *HeraldDB) addAlbum(album Album) (a Album, err error) {
	a, err = hdb.GetUniqueAlbum(album)

	if err != nil && err != ErrNotPresent {
		return Album{}, err
	}

	if a != (Album{}) {
		return a, nil
	}

	// add the album
	query := "INSERT INTO music.albums " +
		"(artist, release_year, n_tracks, n_disks, title, fs_path) " +
		"VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"

	err = hdb.db.QueryRow(query, album.Artist,
		album.Year, album.NumTracks, album.NumDisks,
		album.Title, album.Path).Scan(&album.ID)

	if err != nil {
		return Album{}, err
	}

	return album, nil
}

// addGenre ...
func (hdb *HeraldDB) addGenre(genre Genre) (Genre, error) {
	var err error

	g, err := hdb.GetUniqueGenre(genre)
	if err != nil && err != ErrNotPresent {
		return Genre{}, err
	}

	if g != (Genre{}) {
		return g, nil
	}

	return genre, err
}

// addArtist ...
func (hdb *HeraldDB) addArtist(artist Artist) (a Artist, err error) {
	a, err = hdb.GetUniqueArtist(artist)

	if err != nil && err != ErrNotPresent {
		return Artist{}, err
	}

	if a != (Artist{}) {
		return a, nil
	}

	if !path.IsAbs(artist.Path) {
		return Artist{}, ErrNotAbs
	}

	// add information from artist
	query := "INSERT INTO music.artists (name, fs_path) VALUES ($1, $2) RETURNING id"
	err = hdb.db.QueryRow(query, artist.Name, artist.Path).Scan(&a.ID)

	if err != nil {
		return Artist{}, err
	}

	a.Name = artist.Name
	a.Path = artist.Path

	return a, nil
}

// prepareUniqueQuery ...
func prepareUniqueQuery(table string, rquery reflect.Value) (query string, args []interface{}) {
	rqueryT := rquery.Type()

	selections := make([]string, rquery.NumField())
	args = make([]interface{}, 1)
	args[0] = rquery.FieldByName("ID").Interface()

	for i := 0; i < rquery.NumField(); i++ {
		f := rqueryT.Field(i)
		if tag, ok := f.Tag.Lookup("sql"); ok {
			selections[i] = tag
		}
	}

	query = "SELECT " + strings.Join(selections, ", ") + " " +
		"FROM " + table + " " +
		"WHERE " + "(" + fmt.Sprintf("%s = $%d", "id", 1) + ");"

	return query, args
}

// prepareQuery ...
func prepareQuery(table string, rquery reflect.Value) (query string, args []interface{}) {
	rqueryT := rquery.Type()

	selections := make([]string, rquery.NumField())
	wheres := make([]string, rquery.NumField())
	args = make([]interface{}, rquery.NumField())

	for i := 0; i < rquery.NumField(); i++ {
		f := rqueryT.Field(i)
		if tag, ok := f.Tag.Lookup("sql"); ok {
			selections[i] = tag
			wheres[i] = fmt.Sprintf("%s = $%d", tag, i+1)
			args[i] = rquery.Field(i).Interface()
		}
	}

	query = "SELECT " + strings.Join(selections, ", ") + " " +
		"FROM " + table + " " +
		"WHERE " + "(" + strings.Join(wheres, " AND ") + ");"

	return query, args
}

// prepareDest ...
func prepareDest(rdest *reflect.Value) (destArr []interface{}) {
	destArr = make([]interface{}, rdest.Elem().NumField())
	for i := 0; i < rdest.Elem().NumField(); i++ {
		destArr[i] = rdest.Elem().Field(i).Addr().Interface()
	}
	return destArr
}

// GetUniqueItem ...
// Returns a unique item from the database. Requires an id.
func (hdb *HeraldDB) GetUniqueItem(table string, query interface{}, dest interface{}) (err error) {
	if !isValidTable(table) {
		return ErrInvalidTable
	}
	rquery := reflect.ValueOf(query)
	rdest := reflect.ValueOf(dest)
	if rdest.Kind() != reflect.Ptr || rdest.IsNil() {
		return ErrReflection
	}

	q, a := prepareUniqueQuery(table, rquery)

	destArr := prepareDest(&rdest)

	err = hdb.db.QueryRow(q, a...).Scan(destArr...)
	if err != nil {
		return err
	}

	return nil
}

// GetGenre ...
// Returns the genre matching
func (hdb *HeraldDB) GetGenre(genre Genre) (g Genre, err error) {
	// all genres are unique
	return hdb.GetUniqueGenre(genre)
}

// addImageFile ...
func addImageFile(db *sql.DB, fsPath string) {

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
				err = hdb.processMedia(fsPath, lib)
				if err != nil {
					log.Printf("%v", err)
				}
			}
		case imageType:
			addImageFile(hdb.db, fsPath)
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
