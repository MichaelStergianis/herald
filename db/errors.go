package db

import "errors"

var (
	// ErrNonUnique ...
	// When non unique information is given for a lookup
	ErrNonUnique = errors.New("hdb: information given for query was non-unique")

	// ErrLibAbs ...
	ErrLibAbs = errors.New("hdb: absolute path not given")
)
