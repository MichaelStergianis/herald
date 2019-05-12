package db

import (
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
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// GetTableFromType ...
func GetTableFromType(q interface{}) (string, bool) {
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
	s, ok := validTypes[reflect.TypeOf(qV)]
	return s, ok
}

// GetQueryableFromTable ...
// Get a valid queryable table and a zero value back.
func GetQueryableFromTable(table string) (Queryable, bool) {
	var validTables = map[string]Queryable{
		"music.artists":   &Artist{},
		"music.genres":    &Genre{},
		"music.images":    &Image{},
		"music.albums":    &Album{},
		"music.songs":     &Song{},
		"music.libraries": &Library{},
	}

	q, ok := validTables[table]
	return q, ok
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

// prepareQuery ...
func prepareQuery(table string, rQuery reflect.Value, orderBy []string) (query string, vals []interface{}, err error) {

	rType := rQuery.Type()
	vals = make([]interface{}, 0)
	selectQ := "SELECT "
	fromQ := "FROM " + table + " "
	whereQuery := "WHERE "

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
			// part of the query
			if !IsZero(f) {
				whereQuery += tag + " = " + fmt.Sprintf("$%d", idx) + " "
				vals = append(vals, f.Interface())
				idx++
			}
		}
	}

	if len(vals) < 1 {
		// no where clause necessary if no data provided
		whereQuery = ""
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
	query = selectQ + fromQ + whereQuery + orderQuery + ";"
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

// GetUniqueItem ...
// Returns a unique item from the database. Requires an id.
func (hdb *HeraldDB) GetUniqueItem(query Queryable) (err error) {
	table, ok := GetTableFromType(query)

	if !ok {
		return ErrInvalidTable
	}
	rquery := reflect.ValueOf(query)

	q, a := prepareUniqueQuery(table, rquery)

	destArr := prepareDest(rquery)

	err = hdb.QueryRow(q, a...).Scan(destArr...)
	if err != nil {
		return err
	}

	return nil
}

// GetItem ...
// GetItem searches the database for an item matching the query type,
// using the queries fields.
//
// Order by is optional. You must provide the sql names, you can use
// the provided tag conversion functions to convert from json or
// edn. If you pass an empty array it will be ignored. Otherwise it
// will pass the column names to the sql service.
func (hdb *HeraldDB) GetItem(queryType interface{}, orderBy []string) ([]interface{}, error) {

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

	rows, err := hdb.Query(query, vals...)
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

// addItem ...
// Adds an item to the database. Returning may be the empty string, in
// which case it will return nothing. Otherwise it must be a valid
// interfaceable field for the query type and it will be placed into
// that query and returned.
func (hdb *HeraldDB) addItem(query interface{}, returning []string) (interface{}, error) {
	var err error

	// make a map of sql tags to sql tags to make lookup easy
	returnTags := make(map[string]struct{}, 0)
	for _, ret := range returning {
		returnTags[ret] = struct{}{}
	}

	// lookup corresponding table
	table, ok := GetTableFromType(query)
	if !ok {
		return nil, ErrInvalidTable
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
			insertQ += rType.Field(i).Tag.Get("sql")
			valueQ += fmt.Sprintf("$%d", valNum)
			valNum++
			if i < rQuery.NumField()-1 {
				insertQ += ", "
				valueQ += ", "
			}
			insertVals = append(insertVals, f.Interface())
		}
	}
	insertQ += ") "
	valueQ += ")"

	q := insertQ + valueQ + returningQ
	row := hdb.QueryRow(q, insertVals...)

	if len(returnVal) > 0 {
		err = row.Scan(returnVal...)
	} else {
		err = row.Scan()
	}
	if err != nil {
		return nil, err
	}

	return query, nil
}
