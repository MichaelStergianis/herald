package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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

// isValidTable ...
// Checks to see if the table passed to CountTable is in the list of valid tables.
func isValidTable(table string) bool {
	// create an empty type for our set
	type s struct{}

	var validTables = map[string]s{
		// music schema
		"music.artists":          s{},
		"music.genres":           s{},
		"music.images":           s{},
		"music.albums":           s{},
		"music.images_in_album":  s{},
		"music.songs":            s{},
		"music.libraries":        s{},
		"music.songs_in_library": s{},

		// config schema
		"config.preferences": s{},
		"config.users":       s{},
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

	row := hdb.db.QueryRow(`SELECT COUNT(1) AS count FROM ` + table)

	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CreateLibrary ...
// Creates the library of a given name and path.
func (hdb *HeraldDB) CreateLibrary(name string, path string) {
	stmt, err := hdb.db.Prepare("INSERT INTO music.libraries (library_name, fs_path) VALUES ($1, $2);")
	check(err)
	defer stmt.Close()
	res, err := stmt.Exec(name, path)
	check(err)
	fmt.Println(res)
}

// GetLibraries ...
func (hdb *HeraldDB) GetLibraries() []Library {
	count, err := hdb.CountTable("music.libraries")
	// query
	rows, err := hdb.db.Query("SELECT library_name, fs_path from music.libraries;")
	defer rows.Close()
	check(err)

	libraries := make([]Library, count)
	for rows.Next() {
		var l Library
		err = rows.Scan(&l.Name, &l.Path)
		check(err)
		libraries = append(libraries, l)
	}

	return libraries
}

// GetArtists ...
func (hdb *HeraldDB) GetArtists() []Artist {
	tableName := "music.artists"

	count, err := hdb.CountTable(tableName)

	rows, err := hdb.db.Query("SELECT artist_name from " + tableName)
	defer rows.Close()

	check(err)
	var artists []Artist
	artists = make([]Artist, count)
	for idx := 0; rows.Next(); idx++ {
		var a Artist
		err = rows.Scan(&a.Name)
		check(err)
		artists[idx] = a
	}

	return artists
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

	albm := addAlbum(db, metadata)
	s.Album = &albm

}

// checkSong ...
// Checks to see if the song is already in the database.
func (hdb *HeraldDB) checkSong(song Song) bool {
	hdb.db.QueryRow("select 1 as present from music.songs where fs_path = $1", song.path)

	return false
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

// ScanLibraries ...
// Scans all available libraries
func (hdb *HeraldDB) ScanLibraries() {
	libs := hdb.GetLibraries()
	for _, lib := range libs {
		hdb.ScanLibrary(lib)
	}
}
