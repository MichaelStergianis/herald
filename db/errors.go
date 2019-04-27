package db

import "errors"

var (
	// ErrNotPresent ...
	// Error corresponding to media not being present. Can include
	// albums, songs, artists, images.
	ErrNotPresent = errors.New("hdb: media is not present in database")

	// ErrNonUnique ...
	// When non unique information is given for a lookup
	ErrNonUnique = errors.New("hdb: information given for query was non-unique")

	// ErrNotAbs ...
	ErrNotAbs = errors.New("hdb: absolute path not given")

	// ErrReflection ...
	ErrReflection = errors.New("hdb: unmarshalling encountered a nil value or a non-pointer")

	// ErrInvalidTable ...
	// Returned when an invalid table is requested.
	ErrInvalidTable = errors.New("hdb: invalid table")

	// ErrInvalidTag ...
	// Returned when requesting a field that is not present in a
	// reflected type.
	ErrInvalidTag = errors.New("hdb: invalid tag type")
)
