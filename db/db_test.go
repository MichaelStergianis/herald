package db

import "testing"

// TestCreateDb ...
func TestCreateDb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	hdb := CreateDb()
	if hdb == nil {
		t.Fail()
	}
}
