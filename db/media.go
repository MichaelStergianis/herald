package db

const (
	unknownType = iota
	musicType
	imageType
)

// Library ...
// A representation of a library.
type Library struct {
	ID   int    `edn:":id" json:"id"`
	Name string `edn:":name" json:"name"`
	Path string `edn:":path" json:"path"`
}

// Artist ...
// A representation of an artist.
type Artist struct {
	ID   int    `edn:":id" json:"id"`
	Name string `edn:":name" json:"name"`
}

// Genre ...
// Genre representation.
type Genre struct {
	ID   int    `edn:":id" json:"id"`
	Name string `edn:":name" json:"name"`
}

// Album ...
// Album representation.
type Album struct {
	ID        int `edn:":id" json:"id"`
	Artist    *Artist
	Year      int    `edn:":year" json:"year"`
	NumTracks int    `edn:":num-tracks" json:"num-tracks"`
	Title     string `edn:":title" json:"title"`
	Duration  int    `edn:":duration" json:"duration"` // seconds
}

// Song ...
// Song representation.
type Song struct {
	ID        int `edn:":id" json:"id"`
	Album     *Album
	Path      string `edn:":path" json:"path"`
	Title     string `edn:":title" json:"title"`
	Track     int    `edn:":track" json:"track"`
	NumTracks int    `edn:":num-tracks" json:"num-tracks"`
	Size      int    `edn:":size" json:"size"`         // bytes
	Duration  int    `edn:":duration" json:"duration"` // seconds
	Artist    string `edn:":artist" json:"artist"`
}
