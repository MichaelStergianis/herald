package db

import "testing"

// TestNew ...
// Tests creation of a HeraldDB type.
func TestNew(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	hdb := New()
	if hdb == nil {
		t.Fail()
	}
}
