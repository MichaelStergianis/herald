package db

import (
	"errors"
	"fmt"
)

var (
	// ErrNotPresent ...
	// Error corresponding to media not being present. Can include
	// albums, songs, artists, images.
	ErrNotPresent = errors.New("hdb: media is not present in database")

	// ErrNotAbs ...
	ErrNotAbs = errors.New("hdb: absolute path not given")

	// ErrCannotAddr ...
	ErrCannotAddr = errors.New("hdb: given a value to address that is unaddressable")

	// ErrReflection ...
	ErrReflection = errors.New("hdb: unmarshalling encountered a nil value or a non-pointer")

	// ErrInvalidTable ...
	// Returned when an invalid table is requested.
	ErrInvalidTable = errors.New("hdb: invalid table")

	// ErrInvalidTag ...
	// Returned when requesting a field that is not present in a
	// reflected type.
	ErrInvalidTag = errors.New("hdb: invalid tag type")

	// ErrInvalidScanner ...
	// Returned when given an unkown type to create a sql.Scanner object.
	ErrInvalidScanner = errors.New("hdb: invalid type given to ValueToScanner")
)

// ErrNonUnique ...
// When non unique information is given for a lookup
type ErrNonUnique struct {
	Query interface{}
}

func (e ErrNonUnique) Error() string {
	if e.Query != nil {
		return fmt.Sprintf("hdb: information given for query: %v of type %T was non-unique", e.Query, e.Query)
	}
	return "hdb: information given for query was non-unique"
}
