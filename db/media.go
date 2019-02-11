package db

const (
	unknownType = iota
	musicType
	imageType
)

// Library ...
// A representation of a library.
type Library struct {
	ID   int64  `edn:":id"   json:"id"`
	Name string `edn:":name" json:"name"`
	Path string `edn:":path" json:"path"`
}

// Artist ...
// A representation of an artist.
type Artist struct {
	ID   int64  `edn:":id"   json:"id"`
	Name string `edn:":name" json:"name"`
	Path string `edn:":path" json:"path"`
}

// Genre ...
// Genre representation.
type Genre struct {
	ID   int64  `edn:":id"   json:"id"`
	Name string `edn:":name" json:"name"`
}

// Album ...
// Album representation.
type Album struct {
	ID        int64  `edn:":id"         json:"id"`
	Artist    int64  `edn:":artist"     json:"artist"`
	Year      int    `edn:":year"       json:"year"`
	NumTracks int    `edn:":num-tracks" json:"num-tracks"`
	NumDisks  int    `edn:":num-disks"  json:"num-disks"`
	Title     string `edn:":title"      json:"title"`
	Path      string `edn:":path"       json:"path"`
	Duration  int    `edn:":duration"   json:"duration"` // seconds
}

// Song ...
// Song representation.
type Song struct {
	ID        int64  `edn:":id"         json:"id"`
	Album     int64  `edn:":album"      json:"album"`
	Path      string `edn:":path"       json:"path"`
	Title     string `edn:":title"      json:"title"`
	Track     int    `edn:":track"      json:"track"`
	NumTracks int    `edn:":num-tracks" json:"num-tracks"`
	Disk      int    `edn:":disk"       json:"disk"`
	NumDisks  int    `edn:":num-disks"  json:"num-disks"`
	Size      int    `edn:":size"       json:"size"`     // bytes
	Duration  int    `edn:":duration"   json:"duration"` // seconds
	Artist    string `edn:":artist"     json:"artist"`
}
