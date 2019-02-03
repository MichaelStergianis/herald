package db

import (
	"database/sql"
	"errors"
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

// Close ...
// Closes an hdb.
func (hdb *HeraldDB) Close() {
	hdb.db.Close()
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
func (hdb *HeraldDB) CreateLibrary(name string, path string) (err error) {
	stmt, err := hdb.db.Prepare("INSERT INTO music.libraries (name, fs_path) VALUES ($1, $2);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, path)
	if err != nil {
		return err
	}

	return nil
}

// GetLibraries ...
func (hdb *HeraldDB) GetLibraries() (libs []Library, err error) {
	tableName := "music.libraries"

	count, err := hdb.CountTable(tableName)

	// query
	rows, err := hdb.db.Query("SELECT id, name, fs_path from " + tableName + " ORDER BY id;")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	libs = make([]Library, count)
	for i := 0; rows.Next(); i++ {
		var l Library
		err = rows.Scan(&l.ID, &l.Name, &l.Path)
		if err != nil {
			return nil, err
		}

		libs[i] = l
	}

	return libs, nil
}

// select artists.name, albums.title from music.albums inner join
// music.artists on (music.albums.artist = music.artists.id);

// GetArtists ...
func (hdb *HeraldDB) GetArtists() (artists []Artist, err error) {
	tableName := "music.artists"

	count, err := hdb.CountTable(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := hdb.db.Query("SELECT id, name FROM " + tableName + " ORDER BY id;")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artists = make([]Artist, count)
	for i := 0; rows.Next(); i++ {
		var a Artist
		err = rows.Scan(&a.ID, &a.Name)

		if err != nil {
			return nil, err
		}

		artists[i] = a
	}

	return artists, nil
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
	hdb.db.QueryRow("select 1 as present from music.songs where fs_path = $1", song.Path)

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
	libs, err := hdb.GetLibraries()

	check(err)

	for _, lib := range libs {
		hdb.ScanLibrary(lib)
	}
}
