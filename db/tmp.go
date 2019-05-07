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

// NewFromInterface ...
func NewFromInterface(i interface{}) interface{} {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface()
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
	rdestVal := *rdest
	if rdestVal.Kind() == reflect.Ptr {
		rdestVal = rdestVal.Elem()
	}
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

// GetItem ...
// GetItem searches the database for an item matching the query type,
// using the queries fields.
//
// Order by is optional, if you pass an empty array it will be ignored. Otherwise it will pass the column names to
func (hdb *HeraldDB) GetItem(tableName string, queryType interface{}, converter map[string]string, orderBy []string) ([]interface{}, error) {
	if !GetValidTable(tableName) {
		return nil, ErrInvalidTable
	}

	rQuery := reflect.ValueOf(queryType)
	if rQuery.Kind() == reflect.Ptr {
		rQuery = rQuery.Elem()
	}
	rType := rQuery.Type()

	var (
		vals = make([]interface{}, 0)
	)
	selectQ := "SELECT "
	fromQ := "FROM " + tableName + " "
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

	query := selectQ + fromQ + whereQuery
	if len(orderBy) > 0 {
		orderQuery := "ORDER BY "
		for i, encTag := range orderBy {
			sqlTag, ok := converter[encTag]
			if !ok {
				return []interface{}{}, ErrInvalidTag
			}
			orderQuery += sqlTag
			if i < len(orderBy)-1 {
				orderQuery += ", "
			}
		}
		query += orderQuery
	}
	query += ";"

	rows, err := hdb.Query(query, vals...)
	if err != nil {
		return nil, err
	}

	var results = []interface{}{}
	for rows.Next() {
		r := reflect.New(rType)
		r = r.Elem()

		destArr := prepareDest(&r)

		rows.Scan(destArr...)

		results = append(results, r.Interface())
	}

	return results, nil
}
