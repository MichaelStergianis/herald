package db

import (
	"database/sql"
	"fmt"
	"strconv"
)

const (
	jsonNil = "null"
	ednNil  = "nil"
)

// NullBool ...
type NullBool struct{ sql.NullBool }

// NewNullBool ...
func NewNullBool(b bool) NullBool {
	return NullBool{sql.NullBool{Bool: b, Valid: true}}
}

// MarshalEDN ...
func (v NullBool) MarshalEDN() ([]byte, error) {
	response := ednNil
	if v.Valid {
		response = fmt.Sprintf("%v", v.Bool)
	}
	return []byte(response), nil
}

// MarshalJSON ...
func (v NullBool) MarshalJSON() ([]byte, error) {
	response := jsonNil
	if v.Valid {
		response = fmt.Sprintf("%v", v.Bool)
	}
	return []byte(response), nil
}

// UnmarshalEDN ...
func (v *NullBool) UnmarshalEDN(bytes []byte) error {
	s := string(bytes)
	if s == ednNil {
		v.Valid = false
		return nil
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	v.Bool = b
	v.Valid = true
	return nil
}

// UnmarshalJSON ...
func (v *NullBool) UnmarshalJSON(bytes []byte) error {
	s := string(bytes)
	if s == jsonNil {
		v.Valid = false
		return nil
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	v.Bool = b
	v.Valid = true
	return nil
}

// String ...
func (v NullBool) String() string {
	s := "null"
	if v.Valid {
		s = fmt.Sprintf("%v", v.Bool)
	}
	return s
}

// NullFloat64 ...
type NullFloat64 struct{ sql.NullFloat64 }

// NewNullFloat64 ...
func NewNullFloat64(f float64) NullFloat64 {
	return NullFloat64{sql.NullFloat64{Float64: f, Valid: true}}
}

// MarshalEDN ...
func (v NullFloat64) MarshalEDN() ([]byte, error) {
	response := ednNil
	if v.Valid {
		response = fmt.Sprintf("%v", v.Float64)
	}
	return []byte(response), nil
}

// MarshalJSON ...
func (v NullFloat64) MarshalJSON() ([]byte, error) {
	response := jsonNil
	if v.Valid {
		response = fmt.Sprintf("%v", v.Float64)
	}
	return []byte(response), nil
}

// UnmarshalEDN ...
func (v *NullFloat64) UnmarshalEDN(bytes []byte) error {
	s := string(bytes)
	if s == ednNil {
		v.Valid = false
		return nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	v.Float64 = f
	v.Valid = true
	return nil
}

// UnmarshalJSON ...
func (v *NullFloat64) UnmarshalJSON(bytes []byte) error {
	s := string(bytes)
	if s == jsonNil {
		v.Valid = false
		return nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	v.Float64 = f
	v.Valid = true
	return nil
}

// String ...
func (v NullFloat64) String() string {
	s := "null"
	if v.Valid {
		s = fmt.Sprintf("%v", v.Float64)
	}
	return s
}

// NullInt64 ...
type NullInt64 struct{ sql.NullInt64 }

// NewNullInt64 ...
func NewNullInt64(i int64) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: i, Valid: true}}
}

// MarshalEDN ...
func (v NullInt64) MarshalEDN() ([]byte, error) {
	response := ednNil
	if v.Valid {
		response = fmt.Sprintf("%v", v.Int64)
	}
	return []byte(response), nil
}

// MarshalJSON ...
func (v NullInt64) MarshalJSON() ([]byte, error) {
	response := jsonNil
	if v.Valid {
		response = fmt.Sprintf("%v", v.Int64)
	}
	return []byte(response), nil
}

// UnmarshalEDN ...
func (v *NullInt64) UnmarshalEDN(bytes []byte) error {
	s := string(bytes)
	if s == ednNil {
		v.Valid = false
		return nil
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	v.Int64 = i
	v.Valid = true
	return nil
}

// UnmarshalJSON ...
func (v *NullInt64) UnmarshalJSON(bytes []byte) error {
	s := string(bytes)
	if s == jsonNil {
		v.Valid = false
		return nil
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	v.Int64 = i
	v.Valid = true
	return nil
}

// String ...
func (v NullInt64) String() string {
	s := "null"
	if v.Valid {
		s = fmt.Sprintf("%v", v.Int64)
	}
	return s
}

// NullString ...
type NullString struct{ sql.NullString }

// NewNullString ...
func NewNullString(s string) NullString {
	return NullString{sql.NullString{String: s, Valid: true}}
}

// MarshalEDN ...
func (v NullString) MarshalEDN() ([]byte, error) {
	response := ednNil
	if v.Valid {
		response = v.String
	}
	return []byte(response), nil
}

// MarshalJSON ...
func (v NullString) MarshalJSON() ([]byte, error) {
	response := jsonNil
	if v.Valid {
		response = v.String
	}
	return []byte(response), nil
}

// UnmarshalEDN ...
func (v *NullString) UnmarshalEDN(bytes []byte) error {
	s := string(bytes)

	if s == ednNil {
		v.Valid = false
		return nil
	}

	s, err := strconv.Unquote(s)
	if err != nil {
		return err
	}

	v.String = s
	v.Valid = true
	return nil
}

// UnmarshalJSON ...
func (v *NullString) UnmarshalJSON(bytes []byte) error {
	s := string(bytes)
	if s == jsonNil {
		v.Valid = false
		return nil
	}

	s, err := strconv.Unquote(s)
	if err != nil {
		return err
	}

	v.String = s
	v.Valid = true
	return nil
}
