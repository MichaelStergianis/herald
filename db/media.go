package db

import (
	"reflect"
)

const (
	unknownType = iota
	musicType
	imageType
)

// Queryable ...
type Queryable interface {
	GetID() int64
	SetID(ID int64)
}

func marshall(src interface{}) (dest map[string]interface{}, err error) {
	var rsrc reflect.Value
	if tmp := reflect.ValueOf(src); tmp.Type().String() == "reflect.Value" {
		rsrc = src.(reflect.Value)
	} else {
		rsrc = tmp
	}
	rtype := rsrc.Type()

	dest = make(map[string]interface{}, rsrc.NumField())

	for i := 0; i < rtype.NumField(); i++ {
		f := rtype.Field(i)
		if tag, ok := f.Tag.Lookup("sql"); ok {
			dest[tag] = rsrc.Field(f.Index[0]).Interface()
		}
	}

	return dest, err
}

// unmarshal
func unmarshal(src map[string]interface{}, dest interface{}) (err error) {
	rdest := reflect.ValueOf(dest)
	if rdest.Kind() != reflect.Ptr || rdest.IsNil() {
		return ErrReflection
	}
	rdestT := rdest.Elem().Type()

	for i := 0; i < rdestT.NumField(); i++ {
		f := rdestT.Field(i)
		if tag, ok := f.Tag.Lookup("sql"); ok {
			rdest.Elem().Field(f.Index[0]).Set(reflect.ValueOf(src[tag]))
		}
	}
	return nil
}

// Library ...
// A representation of a library.
type Library struct {
	ID   int64  `edn:"id"   json:"id"   sql:"id"`
	Name string `edn:"name" json:"name" sql:"name"`
	Path string `edn:"-"    json:"-"    sql:"fs_path"`
}

// GetID ...
func (l Library) GetID() int64 {
	return l.ID
}

// SetID ...
func (l *Library) SetID(ID int64) {
	l.ID = ID
}

// Artist ...
// A representation of an artist.
type Artist struct {
	ID   int64  `edn:"id"   json:"id"   sql:"id"`
	Name string `edn:"name" json:"name" sql:"name"`
	Path string `edn:"-"    json:"-"    sql:"fs_path"`
}

// GetID ...
func (a Artist) GetID() int64 {
	return a.ID
}

// SetID ...
func (a *Artist) SetID(ID int64) {
	a.ID = ID
}

// Genre ...
// Genre representation.
type Genre struct {
	ID   int64  `edn:"id"   json:"id"   sql:"id"`
	Name string `edn:"name" json:"name" sql:"name"`
}

// GetID ...
func (g Genre) GetID() int64 {
	return g.ID
}

// SetID ...
func (g *Genre) SetID(ID int64) {
	g.ID = ID
}

// Album ...
// Album representation.
type Album struct {
	ID        int64   `edn:"id"         json:"id"         sql:"id"`
	Artist    int64   `edn:"artist"     json:"artist"     sql:"artist"`
	Year      int     `edn:"year"       json:"year"       sql:"release_year"`
	NumTracks int     `edn:"num-tracks" json:"num-tracks" sql:"num_tracks"`
	NumDisks  int     `edn:"num-disks"  json:"num-disks"  sql:"num_disks"`
	Title     string  `edn:"title"      json:"title"      sql:"title"`
	Path      string  `edn:"-"          json:"-"          sql:"fs_path"`
	Duration  float64 `edn:"duration"   json:"duration"   sql:"duration"` // seconds
}

// GetID ...
func (a Album) GetID() int64 {
	return a.ID
}

// SetID ...
func (a *Album) SetID(ID int64) {
	a.ID = ID
}

// Song ...
// Song representation.
type Song struct {
	ID        int64   `edn:"id"         json:"id"         sql:"id"`
	Album     int64   `edn:"album"      json:"album"      sql:"album"`
	Genre     int64   `edn:"genre"      json:"genre"      sql:"genre"`
	Path      string  `edn:"-"          json:"-"          sql:"fs_path"`
	Title     string  `edn:"title"      json:"title"      sql:"title"`
	Track     int     `edn:"track"      json:"track"      sql:"track"`
	NumTracks int     `edn:"num-tracks" json:"num-tracks" sql:"num_tracks"`
	Disk      int     `edn:"disk"       json:"disk"       sql:"disk"`
	NumDisks  int     `edn:"num-disks"  json:"num-disks"  sql:"num_disks"`
	Size      int     `edn:"size"       json:"size"       sql:"song_size"` // bytes
	Duration  float64 `edn:"duration"   json:"duration"   sql:"duration"`  // seconds
	Artist    string  `edn:"artist"     json:"artist"     sql:"artist"`
}

// GetID ...
func (s Song) GetID() int64 {
	return s.ID
}

// SetID ...
func (s *Song) SetID(ID int64) {
	s.ID = ID
}

// SongInLibrary ...
type SongInLibrary struct {
	SongID    int64 `edn:"song-id" json:"lib-id" sql:"song_id"`
	LibraryID int64 `edn:"lib-id" json:"lib-id" sql:"library_id"`
}

// Image ...
type Image struct {
	ID   int64  `edn:"id"   json:"id"   sql:"id"`
	Path string `edn:"-"    json:"-"    sql:"fs_path"`
}

// GetID ...
func (i Image) GetID() int64 {
	return i.ID
}

// SetID ...
func (i *Image) SetID(ID int64) {
	i.ID = ID
}

// ImageInAlbum ...
type ImageInAlbum struct {
	AlbumID int64 `edn:"album-id" json:"album-id" sql:"album_id"`
	ImageID int64 `edn:"img-id"   json:"img-id"   sql:"image_id"`
}
