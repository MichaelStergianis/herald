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

	"github.com/dhowden/tag"
	// pq is used behind the scenes, but never explicitly used
	_ "github.com/lib/pq"

	ft "gopkg.in/h2non/filetype.v1"
)

// WarblerDB ...
// A type for interfacing with the warbler db
type WarblerDB struct {
	*sql.DB
}

// check ...
func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}

// Open ...
// Creates the connection to the db as a WarblerDB pointer.
func Open(connStr string) (*WarblerDB, error) {
	sqldb, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	wdb := &WarblerDB{
		sqldb,
	}
	return wdb, nil
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
		return d, errors.New("wdb: no duration for empty file")
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
// Closes an wdb.
func (wdb *WarblerDB) Close() {
	wdb.DB.Close()
}

// GetValidTable ...
// Checks to see if the table passed to CountTable is in the list of valid tables.
func GetValidTable(table string) bool {
	// create an empty type for our set
	type empty struct{}
	var validTables = map[string]struct{}{
		// music schema
		"music.artists":   empty{},
		"music.genres":    empty{},
		"music.images":    empty{},
		"music.albums":    empty{},
		"music.songs":     empty{},
		"music.libraries": empty{},

		// config schema
		// "config.preferences": empty{},
		// "config.users":       empty{},

		// Multiple IDs
		"music.images_in_album":  empty{},
		"music.songs_in_library": empty{},
	}

	_, ok := validTables[table]
	return ok
}

// CountTable ...
// Gets the count of a table in our database.
func (wdb *WarblerDB) CountTable(table string) (count int, err error) {
	if ok := GetValidTable(table); !ok {
		return 0, ErrInvalidTable
	}

	row := wdb.QueryRow(`SELECT COUNT(1) AS count FROM ` + table)

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
func (wdb *WarblerDB) AddLibrary(name string, fsPath string) (err error) {
	if !path.IsAbs(fsPath) {
		return ErrNotAbs
	}

	stmt, err := wdb.Prepare("INSERT INTO music.libraries (name, fs_path) VALUES ($1, $2);")
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
func (wdb *WarblerDB) GetLibraries() (libs map[string]Library, err error) {
	tableName := "music.libraries"

	count, err := wdb.CountTable(tableName)

	// query
	rows, err := wdb.Query("SELECT id, name, fs_path from " + tableName + " ORDER BY id;")
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
func (wdb *WarblerDB) processMedia(fsPath string, lib Library) (err error) {
	f, err := os.Open(fsPath)
	if err != nil {
		return err
	}

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		return err
	}

	stats, err := f.Stat()
	if err != nil {
		return err
	}

	s := &Song{
		Path:  fsPath,
		Title: metadata.Title(),
		Size:  stats.Size(),
	}

	// Check to see if the song is in the database
	inLib, err := wdb.songInLibrary(*s, lib)
	if err != nil {
		return err
	}

	if inLib {
		return nil
	}

	t, nT := metadata.Track()
	d, nD := metadata.Disc()
	sqlInts := []*NullInt64{&s.Track, &s.NumTracks, &s.Disk, &s.NumDisks}
	for i, v := range []int{t, nT, d, nD} {
		sqlInts[i].Int64 = int64(v)
		if sqlInts[i].Int64 != 0 {
			sqlInts[i].Valid = true
		}
	}

	s.Duration, err = duration(*s)
	if err != nil {
		return err
	}

	// add genre information
	genre := &Genre{
		Name: metadata.Genre(),
	}
	if genre.Name != "" {
		err = wdb.Create(genre, []string{"id"})
		if err != nil && err != ErrAlreadyExists {
			return err
		}
	}

	// Add the album artist information
	artist := &Artist{
		Name: metadata.AlbumArtist(),
	}
	if artist.Name != "" {
		err = wdb.Create(artist, []string{"id"})
		if err != nil && err != ErrAlreadyExists {
			return err
		}
	}

	albumYear := NewNullInt64(int64(metadata.Year()))
	var albumArtist NullInt64
	if artist.ID != 0 {
		albumArtist.Int64 = artist.ID
		albumArtist.Valid = true
	}

	// Add the album information
	album := &Album{
		Artist:    albumArtist,
		Year:      albumYear,
		NumTracks: s.NumTracks,
		NumDisks:  s.NumDisks,
		Title:     metadata.Album(),
	}

	err = wdb.Create(album, []string{"id"})
	if err != nil && err != ErrAlreadyExists {
		return err
	}

	if album.ID != 0 {
		s.Album = NewNullInt64(album.ID)
	}
	if genre.ID != 0 {
		s.Genre = NewNullInt64(genre.ID)
	}
	for _, v := range []*NullInt64{&s.Album, &s.Genre} {
		if v.Int64 != 0 {
			v.Valid = true
		}
	}

	err = wdb.Create(s, []string{"id"})
	if err != nil && err != ErrAlreadyExists {
		return err
	}

	err = wdb.addSongToLibrary(*s, lib)
	if err != nil {
		return err
	}

	return nil
}

// songInLibrary ...
// Checks to see if a song is in the given library.
func (wdb *WarblerDB) songInLibrary(song Song, library Library) (inLib bool, err error) {
	// get songs id based on path
	if song.Path == "" {
		return false, ErrNonUnique{song}
	}
	err = wdb.QueryRow("SELECT id FROM music.songs where fs_path = $1", song.Path).Scan(&song.ID)
	if err != nil && !(err == sql.ErrNoRows) {
		return false, err
	}

	query := "SELECT COUNT(1) FROM music.songs_in_library where song_id = $1 AND library_id = $2;"

	row := wdb.QueryRow(query, song.ID, library.ID)

	var numInLib int
	err = row.Scan(&numInLib)

	if err != nil {
		return false, err
	}

	if numInLib > 1 {
		return false, ErrNonUnique{library}
	}

	// convert 1/0 to true/false
	inLib = numInLib == 1

	return inLib, nil
}

// TODO GetTypeInLibrary ...

// GetSongsInLibrary ...
func (wdb *WarblerDB) GetSongsInLibrary(lib Library) (songs []Song, err error) {
	if lib.ID == 0 {
		return nil, errors.New("wdb: provided library must have an id")
	}

	var size int
	err = wdb.QueryRow("SELECT COUNT(1) FROM music.songs_in_library WHERE library_id = $1", lib.ID).Scan(&size)
	if err != nil {
		return nil, err
	}
	songs = make([]Song, size)

	rows, err := wdb.Query("SELECT song_id FROM music.songs_in_library WHERE library_id = $1 ORDER BY song_id", lib.ID)
	if err != nil {
		return nil, err
	}
	for idx := 0; rows.Next(); idx++ {
		var s Song
		err = rows.Scan(&s.ID)
		if err != nil {
			return nil, err
		}

		err := wdb.ReadUnique(&s)
		if err != nil {
			return nil, err
		}
		songs[idx] = s
	}

	return songs, nil
}

// addSongToLibrary ...
func (wdb *WarblerDB) addSongToLibrary(song Song, lib Library) error {
	sInL, err := wdb.songInLibrary(song, lib)
	if err != nil {
		return err
	}

	if sInL {
		return nil
	}

	query := "INSERT INTO music.songs_in_library (song_id, library_id) VALUES ($1, $2)"
	stmt, err := wdb.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(&song.ID, &lib.ID)
	if err != nil {
		return err
	}

	return nil
}

// addImageFile ...
func (wdb *WarblerDB) addImageFile(fsPath string) {

}

// ScanLibrary ...
// Scans the library. If some media is already in the library, it will not add it again.
func (wdb *WarblerDB) ScanLibrary(lib Library) (err error) {
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
				err = wdb.processMedia(fsPath, lib)
				if err != nil {
					log.Printf("%v", err)
				}
			}
		case imageType:
			wdb.addImageFile(fsPath)
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
func (wdb *WarblerDB) ScanLibraries() {
	libs, err := wdb.GetLibraries()

	check(err)

	for _, lib := range libs {
		wdb.ScanLibrary(lib)
	}
}

// NewFromQueryable ...
func NewFromQueryable(q Queryable) Queryable {
	t := reflect.TypeOf(q)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface().(Queryable)
}

// NewFromInterface ...
func NewFromInterface(i interface{}) interface{} {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface()
}

// ConvertTags ...
func ConvertTags(tags []string, converter map[string]string) (convertedTags []string, err error) {
	convertedTags = make([]string, len(tags))
	for i, tag := range tags {
		convT, ok := converter[tag]
		if !ok {
			return nil, ErrInvalidTag
		}
		convertedTags[i] = convT
	}
	return convertedTags, nil
}

// NewTagConverter ...
func NewTagConverter(queryType interface{}, from, to string) (converter map[string]string) {
	converter = map[string]string{}
	rValue := reflect.ValueOf(queryType)
	if rValue.Kind() == reflect.Ptr {
		rValue = rValue.Elem()
	}
	rType := rValue.Type()
	// err checking
	for i := 0; i < rType.NumField(); i++ {
		ft := rType.Field(i)
		fv := rValue.Field(i)
		if fv.CanInterface() {
			fromTag, fromOk := ft.Tag.Lookup(from)
			toTag, toOk := ft.Tag.Lookup(to)
			if fromOk && toOk {
				converter[fromTag] = toTag
			}
		}
	}
	return converter
}

// IsZero ...
func IsZero(v reflect.Value) bool {
	t := v.Type()
	var temp reflect.Value
	for temp = v; temp.Kind() == reflect.Ptr; temp = temp.Elem() {
		t = temp.Type()
	}

	n := reflect.Zero(t)

	return temp.Interface() == n.Interface()
}

// GetTableFromType ...
func GetTableFromType(q interface{}) (table string, ok bool) {
	var validTypes = map[reflect.Type]string{
		reflect.TypeOf(&Library{}): "music.libraries",
		reflect.TypeOf(&Artist{}):  "music.artists",
		reflect.TypeOf(&Genre{}):   "music.genres",
		reflect.TypeOf(&Image{}):   "music.images",
		reflect.TypeOf(&Album{}):   "music.albums",
		reflect.TypeOf(&Song{}):    "music.songs",

		reflect.TypeOf(&SongInLibrary{}): "music.songs_in_library",
		reflect.TypeOf(&ImageInAlbum{}):  "music.images_in_album",
	}

	qV := NewFromInterface(q)
	table, ok = validTypes[reflect.TypeOf(qV)]
	return table, ok
}

// prepareDest ...
func prepareDest(rdest reflect.Value) (destArr []interface{}) {
	if rdest.Kind() == reflect.Ptr {
		rdest = rdest.Elem()
	}
	destArr = make([]interface{}, 0)
	for i := 0; i < rdest.NumField(); i++ {
		if rdest.Field(i).CanInterface() {
			destArr = append(destArr, rdest.Field(i).Addr().Interface())
		}
	}
	return destArr
}

// querySelection ...
func querySelection(rQuery reflect.Value) (query string, values []interface{}, err error) {
	if rQuery.Kind() == reflect.Ptr {
		rQuery = rQuery.Elem()
	}
	rType := rQuery.Type()

	query = "SELECT "
	values = make([]interface{}, 0)

	for i := 0; i < rQuery.NumField(); i++ {
		f := rQuery.Field(i)
		if tag, ok := rType.Field(i).Tag.Lookup("sql"); ok {
			if len(values) > 0 {
				query += ", "
			}
			query += tag

			// add values to the respective slices
			if !f.CanAddr() {
				return "", nil, ErrCannotAddr
			}
			values = append(values, f.Addr().Interface())
		}

	}

	return query, values, nil
}

// prepareQuery ...
func prepareQuery(table string, rQuery reflect.Value, orderBy []string) (query string, vals []interface{}, err error) {
	rType := rQuery.Type()
	vals = make([]interface{}, 0)
	selectQ := "SELECT "
	fromQ := "FROM " + table + " "
	whereQ := "WHERE "

	// selection
	idx := 1
	for i := 0; i < rQuery.NumField(); i++ {
		f := rQuery.Field(i)
		if tag, ok := rType.Field(i).Tag.Lookup("sql"); ok {
			// add tag to selection query
			selectQ += tag + " "
			if i < rQuery.NumField()-1 {
				selectQ += ", "
			}

			// if corresponding value is a non zero value, use it as
			// part of the "where query"
			if !IsZero(f) {
				vals = append(vals, f.Interface())
				if idx > 1 {
					whereQ += "AND "
				}
				whereQ += tag + " = " + fmt.Sprintf("$%d", idx) + " "
				idx++
			}
		}
	}
	if len(vals) < 1 {
		// no where clause necessary if no data provided
		whereQ = ""
	}

	orderQuery := ""
	if len(orderBy) > 0 {
		orderQuery += "ORDER BY "
		for i, tag := range orderBy {
			orderQuery += tag
			if i < len(orderBy)-1 {
				orderQuery += ", "
			}
		}
	}
	query = selectQ + fromQ + whereQ + orderQuery + ";"
	return query, vals, nil
}

// prepareUniqueQuery ...
func prepareUniqueQuery(table string, rquery reflect.Value) (query string, args []interface{}) {
	if rquery.Kind() == reflect.Ptr {
		rquery = rquery.Elem()
	}
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
		"WHERE (id = $1);"

	return query, args
}

// ReadUnique ...
// Returns a unique item from the database. Requires an id.
func (wdb *WarblerDB) ReadUnique(query Queryable) (err error) {
	table, ok := GetTableFromType(query)
	if !ok {
		return ErrInvalidTable
	}
	rquery := reflect.ValueOf(query)

	q, a := prepareUniqueQuery(table, rquery)

	destArr := prepareDest(rquery)

	// current issue is that a song has no genre, and we are trying to
	// write <nil> into an int64 space
	// https://stackoverflow.com/questions/28642838/how-do-i-handle-nil-return-values-from-database
	err = wdb.QueryRow(q, a...).Scan(destArr...)
	if err == sql.ErrNoRows {
		return ErrNotPresent
	}
	if err != nil {
		return err
	}

	return nil
}

// Read ...
// Read searches the database for an item matching the query type,
// using the queries fields.
//
// Order by is optional. You must provide the sql names, you can use
// the provided tag conversion functions to convert from json or
// edn. If you pass an empty array it will be ignored. Otherwise it
// will pass the column names to the sql service.
func (wdb *WarblerDB) Read(queryType interface{}, orderBy []string) ([]interface{}, error) {

	table, ok := GetTableFromType(queryType)
	if !ok {
		return nil, ErrInvalidTable
	}

	rQuery := reflect.ValueOf(queryType)
	if rQuery.Kind() == reflect.Ptr {
		rQuery = rQuery.Elem()
	}
	rType := rQuery.Type()

	query, vals, err := prepareQuery(table, rQuery, orderBy)

	rows, err := wdb.Query(query, vals...)
	if err != nil {
		return nil, err
	}

	var results = []interface{}{}
	for rows.Next() {
		r := reflect.New(rType)
		r = r.Elem()

		destArr := prepareDest(r)

		rows.Scan(destArr...)

		results = append(results, r.Interface())
	}

	return results, nil
}

// setMissingValues is a helper function when creating a new row in
// database. It writes any missing values to the supplied value from
// the results. *Mutative function*.
func setMissingValues(src interface{}, dest interface{}) error {
	s := reflect.ValueOf(src)
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr {
		return ErrReflection
	}

	if d.Elem().Type() != s.Type() {
		return ErrTypeMismatch
	}
	d = d.Elem()

	for i := 0; i < s.NumField(); i++ {
		sf := s.Field(i)
		df := d.Field(i)
		if df.CanInterface() && sf.CanInterface() {
			if df.Interface() != sf.Interface() {
				df.Set(sf)
			}
		}
	}

	return nil
}

// Create adds an item to the database. Returning may be the empty
// string, in which case it will return nothing. Otherwise it must be
// a valid interfaceable field for the query type and it will be
// placed into that query and returned.
func (wdb *WarblerDB) Create(query interface{}, returning []string) (err error) {
	// check for existence
	results, err := wdb.Read(query, []string{})
	if err != nil {
		return
	}
	// got more than one result, non unique information provided
	if len(results) > 1 {
		return ErrNonUnique{query}
	}
	// got exactly one, probable match, return
	if len(results) == 1 {
		setMissingValues(results[0], query)
		return ErrAlreadyExists
	}

	// make a map of sql tags to sql tags to make lookup easy
	returnTags := make(map[string]struct{}, 0)
	for _, ret := range returning {
		returnTags[ret] = struct{}{}
	}

	// lookup corresponding table
	table, ok := GetTableFromType(query)
	if !ok {
		return ErrInvalidTable
	}

	rQuery := reflect.ValueOf(query)
	if rQuery.Kind() == reflect.Ptr {
		rQuery = rQuery.Elem()
	}
	rType := rQuery.Type()

	insertVals := make([]interface{}, 0)
	returnVal := make([]interface{}, 0)
	insertQ := "INSERT INTO " + table + " ("
	valueQ := "VALUES ("

	var returningQ string
	if len(returning) > 0 {
		returningQ += " RETURNING "
	}

	valNum := 1
	for i := 0; i < rQuery.NumField(); i++ {
		f := rQuery.Field(i)
		if _, ok := returnTags[rType.Field(i).Tag.Get("sql")]; ok {
			returnVal = append(returnVal, f.Addr().Interface())
			returningQ += rType.Field(i).Tag.Get("sql")
			if len(returnVal) < len(returning) {
				returningQ += ", "
			}
		}
		if !IsZero(f) && f.CanInterface() {
			// insert the field name
			if len(insertVals) > 0 {
				insertQ += ", "
				valueQ += ", "
			}
			insertQ += rType.Field(i).Tag.Get("sql")
			valueQ += fmt.Sprintf("$%d", valNum)
			valNum++
			insertVals = append(insertVals, f.Interface())
		}
	}
	insertQ += ") "
	valueQ += ")"

	q := insertQ + valueQ + returningQ
	row := wdb.QueryRow(q, insertVals...)

	if len(returnVal) > 0 {
		err = row.Scan(returnVal...)
	} else {
		err = row.Scan()
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	if err != nil {
		return
	}

	return nil
}

func setString(statementNum int, set reflect.Value) (int, string, []interface{}) {
	setStr := " SET "
	setVals := make([]interface{}, 0)
	for i := 0; i < set.NumField(); i++ {
		f := set.Field(i)
		if f.CanInterface() && !IsZero(f) {
			tag := set.Type().Field(i).Tag.Get("sql")
			if len(setVals) > 0 {
				setStr += ", "
			}
			setStr += fmt.Sprintf("%s = $%d", tag, statementNum)
			setVals = append(setVals, f.Interface())
			statementNum++
		}
	}
	return statementNum, setStr, setVals
}

// whereString ...
func whereString(statementNum int, where reflect.Value) (int, string, []interface{}) {
	whereStr := " WHERE "
	whereVals := make([]interface{}, 0)

	for i := 0; i < where.NumField(); i++ {
		f := where.Field(i)
		if f.CanInterface() && !IsZero(f) {
			tag := where.Type().Field(i).Tag.Get("sql")
			if len(whereVals) > 0 {
				whereStr += " AND "
			}
			whereStr += fmt.Sprintf("%s = $%d", tag, statementNum)
			whereVals = append(whereVals, f.Interface())
			statementNum++
		}
	}

	return statementNum, whereStr, whereVals
}

// Update ...
func (wdb *WarblerDB) Update(set, where interface{}) (err error) {
	// UPDATE music.songs SET [using `set`] WHERE [using `where`]
	rSet, rWhere := reflect.ValueOf(set), reflect.ValueOf(where)
	if rSet.Kind() == reflect.Ptr {
		rSet = rSet.Elem()
	}
	if rWhere.Kind() == reflect.Ptr {
		rWhere = rWhere.Elem()
	}
	if rSet.Type() != rWhere.Type() {
		return ErrTypeMismatch
	}

	tableName, ok := GetTableFromType(where)
	if !ok {
		return ErrInvalidTable
	}

	updateStr := "UPDATE " + tableName

	var statementNum = 1
	statementNum, setStr, setVals := setString(statementNum, rSet)
	_, whereStr, whereVals := whereString(statementNum, rWhere)

	vals := append(setVals, whereVals...)

	query := updateStr + setStr + whereStr + ";"
	fmt.Printf("%s\n", query)
	fmt.Printf("%#v\n", vals)

	stmt, err := wdb.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(vals...)
	if err != nil {
		return err
	}

	return nil
}
