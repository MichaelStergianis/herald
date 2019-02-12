package db

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
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
		return 0, errors.New("hdb: invalid table to count")
	}

	row := hdb.db.QueryRow(`SELECT COUNT(1) AS count FROM ` + table)

	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// fmtLib ...
// Formats the library string.
func fmtLib(p string) (formatted string) {
	return path.Join(p) + "/"
}

// AddLibrary ...
// Creates the library of a given name and path. Requries an absolute
// path. You should not make assumptions about from which directory
// this server will be run.
func (hdb *HeraldDB) AddLibrary(name string, fsPath string) (err error) {
	fsPath = fmtLib(fsPath)

	if !path.IsAbs(fsPath) {
		return ErrLibAbs
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

// stripLibrary ...
func stripLibrary(fsPath string, lib Library) string {
	return ""
}

// processMedia ...
// Adds a song to the database.
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
	s.Track, s.NumTracks = metadata.Track()
	s.Disk, s.NumDisks = metadata.Disc()

	inLib, err := hdb.songInDatabase(s)
	if err != nil {
		return err
	}

	if inLib {
		return
	}

	artist, err := hdb.addArtist(Artist{
		Name: metadata.AlbumArtist(),
	})
	if err != nil {
		return err
	}

	album, err := hdb.addAlbum(Album{
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

	return nil
}

// songInDatabase ...
// Checks to see if the song is already in the database.
func (hdb *HeraldDB) songInDatabase(song Song) (inLib bool, err error) {
	query := "SELECT COUNT(1) FROM music.songs where songs.fs_path = $1;"

	row := hdb.db.QueryRow(query, song.Path)

	var numInLib int
	err = row.Scan(&numInLib)

	if err != nil {
		return false, err
	}

	if numInLib > 1 {
		return false, errors.New("heraldDB: non-unique row")
	}

	inLib = numInLib == 1

	return inLib, nil
}

// addAlbum ...
// looks in the database for the album information contained in the song metadata,
// if it is not found the function creates and returns the album
func (hdb *HeraldDB) addAlbum(album Album) (a Album, err error) {
	a, err = hdb.GetAlbum(album)

	if err != nil && err != sql.ErrNoRows {
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
		album.Title, album.Path).Scan(&a.ID)

	if err != nil {
		return Album{}, err
	}

	a.Artist = album.Artist
	a.Year = album.Year
	a.NumTracks = album.NumTracks
	a.NumDisks = album.NumDisks
	a.Title = album.Title
	a.Path = album.Path

	return a, nil
}

// addArtist ...
func (hdb *HeraldDB) addArtist(artist Artist) (a Artist, err error) {
	a, err = hdb.GetArtist(artist)

	if err != nil && err != sql.ErrNoRows {
		return Artist{}, err
	}

	if a != (Artist{}) {
		return a, nil
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

// GetArtist ...
func (hdb *HeraldDB) GetArtist(artist Artist) (a Artist, err error) {
	baseQuery := "SELECT id, name, fs_path FROM music.artists WHERE "

	var row *sql.Row

	if artist.ID != 0 {
		row = hdb.db.QueryRow(baseQuery+"artists.id = $1", artist.ID)
	} else if artist.Path != "" {
		row = hdb.db.QueryRow(baseQuery+"artists.fs_path = $1", artist.Path)
	} else {
		return Artist{}, ErrNonUnique
	}

	err = row.Scan(&a.ID, &a.Name, &a.Path)

	if err != nil {
		return Artist{}, err
	}

	return a, nil
}

// GetAlbum ...
// Returns a full album based on some unique information.
// Accepted fields
func (hdb *HeraldDB) GetAlbum(album Album) (a Album, err error) {
	baseQuery := "SELECT id, artist, release_year, n_tracks, n_disks, title, fs_path, duration FROM music.albums WHERE "

	var row *sql.Row

	if album.ID != 0 {
		row = hdb.db.QueryRow(baseQuery+"albums.id = $1", album.ID)
	} else if album.Path != "" {
		row = hdb.db.QueryRow(baseQuery+"albums.fs_path = $1", album.Path)
	} else {
		return Album{}, ErrNonUnique
	}

	err = row.Scan(&a.ID, &a.Artist, &a.Year, &a.NumTracks, &a.NumDisks, &a.Title, &a.Path, &a.Duration)

	// error scanning the row
	if err != nil {
		return Album{}, err
	}

	return a, nil
}

// addImageFile ...
func addImageFile(db *sql.DB, fsPath string) {

}

// ScanLibrary ...
// Scans the library. If some media is already in the library, it will not add it again.
func (hdb *HeraldDB) ScanLibrary(lib Library) {
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
			hdb.processMedia(fsPath, lib)
		case imageType:
			addImageFile(hdb.db, fsPath)
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
