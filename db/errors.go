package db

import (
	"errors"
	"fmt"
)

var (
	// ErrNotPresent corresponds to media not being present. Can include
	// albums, songs, artists, images.
	ErrNotPresent = errors.New("wdb: media is not present in database")

	// ErrAlreadyExists occurs during Create when the supplied value already existed in the database
	ErrAlreadyExists = errors.New("wdb: attempt to create value failed; value already exists")

	// ErrNotAbs ...
	ErrNotAbs = errors.New("wdb: absolute path not given")

	// ErrCannotAddr ...
	ErrCannotAddr = errors.New("wdb: given a value to address that is unaddressable")

	// ErrReflection ...
	ErrReflection = errors.New("wdb: unmarshalling encountered a nil value or a non-pointer")

	// ErrTypeMismatch is returned when a type mismatch occurs during reflection.
	ErrTypeMismatch = errors.New("wdb: type mismatch")

	// ErrInvalidTable is returned when an invalid table is requested.
	ErrInvalidTable = errors.New("wdb: invalid table")

	// ErrInvalidTag is returned when requesting a field that is not
	// present in a reflected type.
	ErrInvalidTag = errors.New("wdb: invalid tag type")

	// ErrInvalidScanner is returned when given an unkown type to
	// create a sql.Scanner object.
	ErrInvalidScanner = errors.New("wdb: invalid type given to ValueToScanner")
)

// ErrNonUnique occurs When non unique information is given for a
// lookup
type ErrNonUnique struct {
	Query interface{}
}

func (e ErrNonUnique) Error() string {
	if e.Query != nil {
		return fmt.Sprintf("wdb: information given for query: %v of type %T was non-unique", e.Query, e.Query)
	}
	return "wdb: information given for query was non-unique"
}
