package db

import (
	"log"
	"os"
	"testing"

	testfixtures "gopkg.in/testfixtures.v2"
)

var (
	dbName   string
	hdb      *HeraldDB
	fixtures *testfixtures.Context
)

// paramstoStr ...
// Takes a map of parameters and generates a string.
func paramstoStr(params map[string]string) (s string) {
	s = ""
	for k, v := range params {
		s = s + k + "=" + v + " "
	}
	return s
}

// prepareTestDatabase ...
func prepareTestDatabase() {
	if err := fixtures.Load(); err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error

	hdb, err = Open("dbname=herald_test user=herald sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	fixtures, err = testfixtures.NewFolder(hdb.db, &testfixtures.PostgreSQL{UseAlterConstraint: true}, "fixtures")
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// TestCountTable ...
func TestCountTable(t *testing.T) {
	prepareTestDatabase()
	count, err := hdb.CountTable("music.artists")

	if err != nil {
		t.Error(err)
	}

	if count != 4 {
		t.Fail()
	}
}
