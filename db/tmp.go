package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

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
// the results.
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

// Create ...
// Adds an item to the database. Returning may be the empty string, in
// which case it will return nothing. Otherwise it must be a valid
// interfaceable field for the query type and it will be placed into
// that query and returned.
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
