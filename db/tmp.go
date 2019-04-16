package db

import (
	"fmt"
	"reflect"
	"strings"
)

// GetQueryable ...
// Get a valid queryable table and a zero value back.
func GetQueryable(table string) (Queryable, bool) {

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

// GetUniqueItem ...
// Returns a unique item from the database. Requires an id.
func (hdb *HeraldDB) GetUniqueItem(table string, query Queryable) (err error) {
	if _, ok := GetQueryable(table); !ok {
		return ErrInvalidTable
	}
	rquery := reflect.ValueOf(query)

	q, a := prepareUniqueQuery(table, rquery)

	destArr := prepareDest(&rquery)

	err = hdb.QueryRow(q, a...).Scan(destArr...)
	if err != nil {
		return err
	}

	return nil
}

// NewFromQueryable ...
func NewFromQueryable(q Queryable) Queryable {
	t := reflect.TypeOf(q)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface().(Queryable)
}

// prepareQuery ...
func prepareQuery(table string, rquery reflect.Value) (query string, args []interface{}) {
	rqueryT := rquery.Type()

	selections := make([]string, rquery.NumField())
	wheres := make([]string, rquery.NumField())
	args = make([]interface{}, rquery.NumField())

	for i := 0; i < rquery.NumField(); i++ {
		f := rqueryT.Field(i)
		if tag, ok := f.Tag.Lookup("sql"); ok {
			selections[i] = tag
			wheres[i] = fmt.Sprintf("%s = $%d", tag, i+1)
			args[i] = rquery.Field(i).Interface()
		}
	}

	query = "SELECT " + strings.Join(selections, ", ") + " " +
		"FROM " + table + " " +
		"WHERE " + "(" + strings.Join(wheres, " AND ") + ");"

	return query, args
}

// prepareDest ...
func prepareDest(rdest *reflect.Value) (destArr []interface{}) {
	rdestVal := rdest.Elem()
	destArr = make([]interface{}, rdestVal.NumField())
	for i := 0; i < rdestVal.NumField(); i++ {
		if rdestVal.Field(i).CanInterface() {
			destArr[i] = rdestVal.Field(i).Addr().Interface()
		}
	}
	return destArr
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
		"WHERE " + "(" + fmt.Sprintf("%s = $%d", "id", 1) + ");"

	return query, args
}
