package db

import "database/sql"

// GetUniqueArtist ...
func (hdb *HeraldDB) GetUniqueArtist(artist Artist) (a Artist, err error) {
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

	if err == sql.ErrNoRows {
		return Artist{}, ErrNotPresent
	}

	if err != nil {
		return Artist{}, err
	}

	return a, nil
}

// GetUniqueGenre ...
func (hdb *HeraldDB) GetUniqueGenre(genre Genre) (g Genre, err error) {
	baseQuery := "SELECT id, name " +
		"FROM music.genres WHERE "

	var row *sql.Row

	if genre.ID != 0 {
		row = hdb.db.QueryRow(baseQuery+"genres.id = $1", genre.ID)
	} else if genre.Name != "" {
		row = hdb.db.QueryRow(baseQuery+"genres.name = $1", genre.Name)
	} else {
		return Genre{}, ErrNonUnique
	}

	err = row.Scan(
		&g.ID,
		&g.Name,
	)

	if err == sql.ErrNoRows {
		return Genre{}, ErrNotPresent
	}

	if err != nil {
		return Genre{}, err
	}

	return g, nil
}

// GetUniqueSong ...
func (hdb *HeraldDB) GetUniqueSong(song Song) (s Song, err error) {
	baseQuery := "SELECT id, album, genre, fs_path, title, track, num_tracks, disk, num_disks, song_size, duration, artist " +
		"FROM music.songs WHERE "

	var row *sql.Row

	if song.ID != 0 {
		row = hdb.db.QueryRow(baseQuery+"songs.id = $1", song.ID)
	} else if song.Path != "" {
		row = hdb.db.QueryRow(baseQuery+"songs.fs_path = $1", song.Path)
	} else {
		return Song{}, ErrNonUnique
	}

	err = row.Scan(
		&s.ID,
		&s.Album,
		&s.Genre,
		&s.Path,
		&s.Title,
		&s.Track,
		&s.NumTracks,
		&s.Disk,
		&s.NumDisks,
		&s.Size,
		&s.Duration,
		&s.Artist,
	)

	if err == sql.ErrNoRows {
		return Song{}, ErrNotPresent
	}

	if err != nil {
		return Song{}, err
	}

	return s, nil
}

// GetUniqueAlbum ...
// Returns a full album based on some unique information.
// Accepted fields
func (hdb *HeraldDB) GetUniqueAlbum(album Album) (a Album, err error) {
	baseQuery := "SELECT id, artist, release_year, n_tracks, n_disks, title, fs_path, duration FROM music.albums WHERE "

	var row *sql.Row

	if album.ID != 0 {
		row = hdb.db.QueryRow(baseQuery+"albums.id = $1", album.ID)
	} else if album.Path != "" {
		row = hdb.db.QueryRow(baseQuery+"albums.fs_path = $1", album.Path)
	} else {
		return Album{}, ErrNonUnique
	}

	err = row.Scan(
		&a.ID,
		&a.Artist,
		&a.Year,
		&a.NumTracks,
		&a.NumDisks,
		&a.Title,
		&a.Path,
		&a.Duration,
	)

	if err == sql.ErrNoRows {
		return Album{}, ErrNotPresent
	}

	// error scanning the row
	if err != nil {
		return Album{}, err
	}

	return a, nil
}
