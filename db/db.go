package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dhowden/tag"

	// pq uses golang sql
	_ "github.com/lib/pq"
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

const (
	unknownType = iota
	musicType
	imageType
)

type duration time.Duration

// Library ...
// A representation of a library.
type Library struct {
	ID   int
	Name string
	Path string
}

// Artist ...
// A representation of an artist.
type Artist struct {
	Name string
}

// Album ...
// Album representation.
type Album struct {
	Artist    Artist
	AlbumSize int
	Title     string
	Duration  duration
}

// Song ...
// Song representation.
type Song struct {
	Title     string
	Album     Album
	Track     int
	NumTracks int
}

// isValidTable ...
// Checks to see if the table passed to CountTable is in the list of valid tables.
func isValidTable(table string) bool {
	var validTables = map[string]bool{
		"music.artists":          true,
		"music.genres":           true,
		"music.images":           true,
		"music.albums":           true,
		"music.images_in_album":  true,
		"music.songs":            true,
		"music.libraries":        true,
		"music.songs_in_library": true,
	}

	_, inTable := validTables[table]
	return inTable
}

// CountTable ...
// Gets the count of a table in our database.
func (hdb *HeraldDB) CountTable(table string) (count int, err error) {
	if !isValidTable(table) {
		return 0, errors.New("In countTable: Invalid table")
	}

	rows, err := hdb.db.Query(`SELECT COUNT(*) AS count FROM ` + table)
	check(err)

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&count)
		check(err)
	}
	return count, nil
}

// New ...
// creates the connection to the db as a HeraldDB pointer.
func New() *HeraldDB {
	connStr := "user=herald dbname=herald_db sslmode=disable"
	sqldb, err := sql.Open("postgres", connStr)
	check(err)
	hdb := HeraldDB{
		sqldb,
	}
	return &hdb
}

// CreateLibrary ...
// Creates the library of a given name and path.
func (hdb *HeraldDB) createLibrary(name string, path string) {
	stmt, err := hdb.db.Prepare("INSERT INTO music.libraries (library_name, fs_path) VALUES ($1, $2);")
	check(err)
	defer stmt.Close()
	res, err := stmt.Exec(name, path)
	check(err)
	fmt.Println(res)
}

// GetLibraries ...
func (hdb *HeraldDB) GetLibraries() []Library {
	// query
	rows, err := hdb.db.Query("SELECT library_name, fs_path from music.libraries;")
	defer rows.Close()

	var libraries []Library
	for rows.Next() {
		var l Library
		err = rows.Scan(&l.Name, &l.Path)
		check(err)
		libraries = append(libraries, l)
	}

	return libraries
}

// checkFileType ...
func checkFileType(file string) int {
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

// addSong ...
// Adds a song to the database.
func addSong(db *sql.DB, path string) {
	f, err := os.Open(path)
	check(err)

	metadata, err := tag.ReadFrom(f)
	check(err)

	var s Song
	s.Title = metadata.Title()
	s.Track, s.NumTracks = metadata.Track()

	s.Album = addAlbum(db, metadata)

}

// checkAlbum ...
// checks if the album exists in the database
func checkAlbum(db *sql.DB, albm Album) (isIn bool) {
	// stmt, err := db.Prepare("SELECT COUNT(*) FROM music.albums WHERE music.albums.title = ?;")
	// check(err)

	return false
}

// addAlbum ...
// looks in the database for the album information contained in the song metadata,
// if it is not found the function creates and returns the album
func addAlbum(db *sql.DB, metadata tag.Metadata) Album {
	// albumName := metadata.Album()
	// _, numTracks := metadata.Track()
	// albumArtist := metadata.AlbumArtist()

	return Album{}
}

// addImageFile ...
func addImageFile(db *sql.DB, path string) {

}

// ScanLibrary ...
// Scans the library. If some media is already in the library, it will not add it again.
func (hdb *HeraldDB) ScanLibrary(lib Library) {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Encountered the following error while traversing %q: %v", path, err)
			return err
		}
		if info.IsDir() {
			return err
		}
		switch checkFileType(path) {
		case musicType:
			addSong(hdb.db, path)
		case imageType:
			addImageFile(hdb.db, path)
		}
		return err
	}
	filepath.Walk(lib.Path, walkFn)
}
