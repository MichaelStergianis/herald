package db

import (
	"log"

	"gopkg.in/testfixtures.v2"
)

// PrepareTestDatabase ...
func PrepareTestDatabase(wdb *WarblerDB, fixturesDir string) (func(), error) {
	var err error

	fixtures, err := testfixtures.NewFolder(wdb.DB, &testfixtures.PostgreSQL{UseAlterConstraint: true}, fixturesDir)
	if err != nil {
		return nil, err
	}

	f := func() {
		if err := fixtures.Load(); err != nil {
			log.Fatal(err)
		}
	}

	return f, nil
}
