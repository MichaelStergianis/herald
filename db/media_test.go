package db

import (
	"testing"
)

// TestMarshall ...
func TestMarshall(t *testing.T) {
	lib := Library{
		ID:   1,
		Name: "MyLib",
		Path: "/home/test/MyLib",
	}

	verificationLib := map[string]interface{}{
		"id":      int64(1),
		"name":    "MyLib",
		"fs_path": "/home/test/MyLib",
	}

	m, err := marshall(lib)

	if err != nil {
		t.Error(err)
	}

	for k, v := range m {
		if v != verificationLib[k] {
			t.Errorf("unexpected result: expected: %v result: %v\n", verificationLib[k], v)
		}
	}

}

// TestUnmarshal ...
func TestUnmarshal(t *testing.T) {
	m := map[string]interface{}{
		"id":      int64(1),
		"name":    "MyLib",
		"fs_path": "/home/test/MyLib",
	}

	verificationLib := Library{
		ID:   1,
		Name: "MyLib",
		Path: "/home/test/MyLib",
	}

	var lib Library

	err := unmarshal(m, &lib)
	if err != nil {
		t.Error(err)
	}

	if lib.ID != verificationLib.ID {
		t.Error("id result does not match")
	}

	if lib.Name != verificationLib.Name {
		t.Error("name result does not match")
	}

	if lib.Path != verificationLib.Path {
		t.Error("path result does not match")
	}

	err = unmarshal(m, lib)
	if err == nil {
		t.Error("testing negative case of unmarshal, did not receive error")
	}

	if err != ErrReflection {
		t.Errorf("expected reflection error, got %v\n", err)
	}

}
