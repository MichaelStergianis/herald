package db

import (
	"database/sql"
	"encoding/json"
	"testing"

	"gopkg.in/edn.v1"
)

var (
	encoders = []func(interface{}) ([]byte, error){
		edn.Marshal,
		json.Marshal,
	}

	decoders = []struct {
		f    func([]byte, interface{}) error
		name string
	}{
		{edn.Unmarshal, "edn"},
		{json.Unmarshal, "json"},
	}
)

// NullValue ...
// NullValue describes an all-in-one for testing purposes.
type NullValue interface {
	// marshalers
	json.Marshaler
	edn.Marshaler
}

// TestMarshal ...
func TestMarshal(t *testing.T) {
	type Container struct {
		Value NullValue `json:"value" edn:"value"`
	}

	cases := [...]struct {
		container Container
		answer    []string
	}{
		// bool
		{Container{Value: &NullBool{sql.NullBool{Bool: true, Valid: true}}}, []string{"{:value true}", `{"value":true}`}},
		{Container{Value: &NullBool{sql.NullBool{}}}, []string{"{:value nil}", `{"value":null}`}},

		// int
		{Container{Value: &NullInt64{sql.NullInt64{Int64: 4, Valid: true}}}, []string{"{:value 4}", `{"value":4}`}},
		{Container{Value: &NullInt64{sql.NullInt64{}}}, []string{"{:value nil}", `{"value":null}`}},
	}

	for i, encoder := range encoders {
		for testCase, test := range cases {
			b, err := encoder(test.container)
			if err != nil {
				t.Error(err)
			}

			if string(b) != test.answer[i] {
				t.Errorf("case %d failed:\n\texpected: %v\n\tresult: %v", testCase, test.answer[i], string(b))
			}

		}
	}
}

// TestUnmarshalNullBool ...
func TestUnmarshalNullBool(t *testing.T) {
	type Container struct {
		Value NullBool `json:"value" edn:"value"`
	}
	cases := []struct {
		queries []string
		holder  *Container
		answer  *Container
	}{
		{[]string{"{:value false}", `{"value": false}`},
			&Container{Value: NullBool{}},
			&Container{Value: NullBool{sql.NullBool{false, true}}},
		},

		{[]string{"{:value true}", `{"value": true}`},
			&Container{Value: NullBool{}},
			&Container{Value: NullBool{sql.NullBool{true, true}}},
		},

		// null cases
		{[]string{"{:value nil}", `{"value": null}`},
			&Container{Value: NullBool{}},
			&Container{Value: NullBool{sql.NullBool{false, false}}},
		},

		{[]string{"{}", `{}`},
			&Container{Value: NullBool{}},
			&Container{Value: NullBool{}},
		},
	}

	for testCase, test := range cases {
		for i, decoder := range decoders {

			err := decoder.f([]byte(test.queries[i]), test.holder)
			if err != nil {
				t.Errorf("error in test case %d: %v", testCase, err)
			}

			if *test.answer != *test.holder {
				t.Errorf("test case: %d\n\tdecoder: %s\n\tquery: %s\n\texpected: %+v\n\tresult:  %+v\n",
					testCase, decoder.name, test.queries[i], *test.answer, *test.holder)
			}

			// reset test holder
			test.holder = &Container{Value: NullBool{}}
		}
	}
}

// TestUnmarshalNullFloat64 ...
func TestUnmarshalNullFloat64(t *testing.T) {
	type Container struct {
		Value NullFloat64 `json:"value" edn:"value"`
	}
	cases := []struct {
		queries []string
		holder  *Container
		answer  *Container
	}{
		// 0
		{[]string{"{:value 3.14}", `{"value": 3.14}`},
			&Container{Value: NullFloat64{}},
			&Container{Value: NullFloat64{sql.NullFloat64{3.14, true}}},
		},

		{[]string{"{:value 0}", `{"value": 0}`},
			&Container{Value: NullFloat64{}},
			&Container{Value: NullFloat64{sql.NullFloat64{0.0, true}}},
		},

		{[]string{"{:value -3.14}", `{"value": -3.14}`},
			&Container{Value: NullFloat64{}},
			&Container{Value: NullFloat64{sql.NullFloat64{-3.14, true}}},
		},

		// null cases
		{[]string{"{:value nil}", `{"value": null}`},
			&Container{Value: NullFloat64{}},
			&Container{Value: NullFloat64{sql.NullFloat64{0.0, false}}},
		},

		{[]string{"{}", `{}`},
			&Container{Value: NullFloat64{}},
			&Container{Value: NullFloat64{}},
		},
	}

	for testCase, test := range cases {
		for i, decoder := range decoders {

			err := decoder.f([]byte(test.queries[i]), test.holder)
			if err != nil {
				t.Errorf("error in test case %d: %v", testCase, err)
			}

			if *test.answer != *test.holder {
				t.Errorf("test case: %d\n\tdecoder: %s\n\tquery: %s\n\texpected: %+v\n\tresult:  %+v\n",
					testCase, decoder.name, test.queries[i], *test.answer, *test.holder)
			}

			// reset test holder
			test.holder = &Container{Value: NullFloat64{}}
		}
	}
}

// TestUnmarshalNullInt64 ...
func TestUnmarshalNullInt64(t *testing.T) {

	type Container struct {
		Value NullInt64 `json:"value" edn:"value"`
	}
	cases := []struct {
		queries []string
		holder  *Container
		answer  *Container
	}{
		// valued cases
		{[]string{"{:value 4}", `{"value":4}`},
			&Container{Value: NullInt64{}},
			&Container{Value: NullInt64{sql.NullInt64{4, true}}}},

		{[]string{"{:value 0}", `{"value":0}`},
			&Container{Value: NullInt64{}},
			&Container{Value: NullInt64{sql.NullInt64{0, true}}}},

		{[]string{"{:value -4}", `{"value":-4}`},
			&Container{Value: NullInt64{}},
			&Container{Value: NullInt64{sql.NullInt64{-4, true}}}},

		// null cases
		{[]string{"{:value nil}", `{"value": null}`},
			&Container{Value: NullInt64{}},
			&Container{Value: NullInt64{sql.NullInt64{0, false}}}},

		{[]string{"{}", `{}`},
			&Container{Value: NullInt64{}},
			&Container{Value: NullInt64{}}},
	}

	for testCase, test := range cases {
		for i, decoder := range decoders {

			err := decoder.f([]byte(test.queries[i]), test.holder)
			if err != nil {
				t.Errorf("error in test case %d: %v", testCase, err)
			}

			if *test.answer != *test.holder {
				t.Errorf("test case: %d\n\tdecoder: %s\n\tquery: %s\n\texpected: %+v\n\tresult:  %+v\n",
					testCase, decoder.name, test.queries[i], *test.answer, *test.holder)
			}

			// reset test holder
			test.holder = &Container{Value: NullInt64{}}

		}
	}

}

// TestUnmarshalNullString ...
func TestUnmarshalNullString(t *testing.T) {

	type Container struct {
		Value NullString `json:"value" edn:"value"`
	}
	cases := []struct {
		queries []string
		holder  *Container
		answer  *Container
	}{
		// valued cases
		{[]string{`{:value "hello"}`, `{"value":"hello"}`},
			&Container{Value: NullString{}},
			&Container{Value: NullString{sql.NullString{"hello", true}}}},

		{[]string{`{:value "0"}`, `{"value":"0"}`},
			&Container{Value: NullString{}},
			&Container{Value: NullString{sql.NullString{"0", true}}}},

		{[]string{`{:value "world"}`, `{"value":"world"}`},
			&Container{Value: NullString{}},
			&Container{Value: NullString{sql.NullString{"world", true}}}},

		// null cases
		{[]string{`{:value nil}`, `{"value": null}`},
			&Container{Value: NullString{}},
			&Container{Value: NullString{sql.NullString{"", false}}}},

		{[]string{`{}`, `{}`},
			&Container{Value: NullString{}},
			&Container{Value: NullString{}}},
	}

	for testCase, test := range cases {
		for i, decoder := range decoders {

			err := decoder.f([]byte(test.queries[i]), test.holder)
			if err != nil {
				t.Errorf("error in test case %d: %v", testCase, err)
			}

			if *test.answer != *test.holder {
				t.Errorf("test case: %d\n\tdecoder: %s\n\tquery: %s\n\texpected: %+v\n\tresult:  %+v\n",
					testCase, decoder.name, test.queries[i], *test.answer, *test.holder)
			}

			// reset test holder
			test.holder = &Container{Value: NullString{}}

		}
	}

}
