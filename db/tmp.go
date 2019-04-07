package db

import "reflect"

// prepareDest ...
func prepareDest(rdest *reflect.Value) (destArr []interface{}) {
	rdestVal := rdest.Elem()
	destArr = make([]interface{}, rdestVal.NumField())
	for i := 0; i < rdestVal.NumField(); i++ {
		destArr[i] = rdestVal.Field(i).Addr().Interface()
	}
	return destArr
}

// GetUniqueItem ...
// Returns a unique item from the database. Requires an id.
func (hdb *HeraldDB) GetUniqueItem(table string, query interface{}, dest interface{}) (err error) {
	if _, ok := GetValidTable(table); !ok {
		return ErrInvalidTable
	}
	rquery := reflect.ValueOf(query)
	rdest := reflect.ValueOf(dest)
	if rdest.Kind() != reflect.Ptr || rdest.IsNil() {
		return ErrReflection
	}

	q, a := prepareUniqueQuery(table, rquery)

	destArr := prepareDest(&rdest)

	err = hdb.QueryRow(q, a...).Scan(destArr...)
	if err != nil {
		return err
	}

	return nil
}
