package main

import (
	"database/sql"
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

const (
	unknownType = iota
	musicType
	imageType
)

type duration time.Duration

type library struct {
	name string
	path string
}

type artist struct {
	name string
}

type album struct {
	artist    artist
	albumSize int
	title     string
	duration  duration
}

type song struct {
	title     string
	album     album
	track     int
	numTracks int
}

// getCount ...
// gets a count of an arbitrary table in our db
func getCount(db *sql.DB, tableName string) (count int) {
	stmt, err := db.Prepare("SELECT COUNT(*) as count FROM ?")
	check(err)

	rows, err := stmt.Query(tableName)
	check(err)
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&count)
		check(err)
	}
	return count
}

// createDb ...
// creates the connection to the db
func createDb() *sql.DB {
	connStr := "user=herald dbname=herald_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	check(err)
	return db
}

// initDB ...
// checks whether or not the database exists given the db
func initDB(db *sql.DB) {

}

// createLibrary ...
func createLibrary(db *sql.DB, name string, path string) {
	stmt, err := db.Prepare("INSERT INTO libraries (library_name, fs_path) VALUES ($1, $2);")
	check(err)
	defer stmt.Close()
	res, err := stmt.Exec(name, path)
	check(err)
	fmt.Println(res)
}

// getLibraries ...
func getLibraries(db *sql.DB) []library {
	// query
	rows, err := db.Query(`SELECT library_name, fs_path from libraries;`)
	defer rows.Close()

	var libraries []library
	for rows.Next() {
		var l library
		err = rows.Scan(&l.name, &l.path)
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

// addMusicFile ...
func addMusicFile(db *sql.DB, path string) {
	f, err := os.Open(path)
	check(err)

	metadata, err := tag.ReadFrom(f)
	check(err)

	var s song
	s.title = metadata.Title()
	s.track, s.numTracks = metadata.Track()

	s.album = addAlbum(db, metadata)

}

// checkAlbum ...
// checks if the album exists in the database
func checkAlbum(db *sql.DB, albm album) (isIn bool) {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM ")
	check(err)

}

// addAlbum ...
// looks in the database for the album information contained in the song metadata,
// if it is not found the function creates and returns the album
func addAlbum(db *sql.DB, metadata tag.Metadata) album {
	albumName := metadata.Album()
	_, numTracks := metadata.Track()
	albumArtist := metadata.AlbumArtist()

	return album{}
}

// addImageFile ...
func addImageFile(db *sql.DB, path string) {

}

func scanLibrary(db *sql.DB, lib library) {
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
			addMusicFile(db, path)
		case imageType:
			addImageFile(db, path)
		}
		return err
	}
	filepath.Walk(lib.path, walkFn)
}
