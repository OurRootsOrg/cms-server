package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// fieldType represents the type of an attribute value
type fieldType string

const (
	// IntType indicates a field of type int
	IntType fieldType = "Int"
	// StringType indicates a field of type string
	StringType = "String"
	// ImageType indicates a field of type image.Image
	ImageType = "Image"
	// LocationType indicates a field of type Location
	LocationType = "Location"
	// TimeType indicates a field of type time.Time
	TimeType = "Time"
)

// fieldTypes returns a map of the valid field types
func fieldTypes() map[fieldType]bool {
	return map[fieldType]bool{IntType: true, StringType: true, ImageType: true, LocationType: true, TimeType: true}
}

// makeFieldType makes a fieldType from a string, but returns an error if the string isn't a valid type
func isValidFieldType(ft fieldType) error {
	if fieldTypes()[ft] {
		return nil
	}
	return errors.New("'" + string(ft) + "' is not a valid fieldType")
}

func makeFieldType(s string) (fieldType, error) {
	ft := fieldType(s)
	err := isValidFieldType(ft)
	if err != nil {
		return fieldType(""), err
	}
	return ft, nil
}

// FieldDef defines a name and type of a field
type FieldDef struct {
	name      string
	fieldType fieldType
}

// NewFieldDef constructs a FieldDef
func NewFieldDef(name string, fieldType string) (FieldDef, error) {
	ft, err := makeFieldType(fieldType)
	if err != nil {
		return FieldDef{}, err
	}
	return FieldDef{name: name, fieldType: ft}, nil
}

// Name returns the name of a field
func (f FieldDef) Name() string {
	return f.name
}

// Type returns the type of a StringDef
func (f FieldDef) Type() string {
	return string(f.fieldType)
}

// StringField represent a string field
type StringField struct {
	FieldDef
	value string
}

// AsString returns the field value as a string
func (s *StringField) AsString() string {
	return s.value
}

// SetString set the field value from a string
func (s *StringField) SetString(value string) {
	s.value = value
}

// IntField represent an int64 field
type IntField struct {
	FieldDef
	value int64
}

// AsInt64 returns the field value as an int64
func (s *IntField) AsInt64() int64 {
	return s.value
}

// SetInt64 set the field value from an int64
func (s *IntField) SetInt64(value int64) {
	s.value = value
}

// TimeField represent a string field
type TimeField struct {
	FieldDef
	value time.Time
}

// AsTime returns the field value as a time.Time
func (s *TimeField) AsTime() time.Time {
	return s.value
}

// SetTime set the field value from a time.Time
func (s *TimeField) SetTime(value time.Time) {
	s.value = value
}

// LocationField represent a string field
type LocationField struct {
	FieldDef
	value    string
	rawValue string
}

// AsString returns the field value as a string
func (s *LocationField) AsString() string {
	return s.value
}

// SetString set the field value from a string
func (s *LocationField) SetString(value string) {
	s.rawValue = value
	// Do some canonicalization
	s.value = value
}

// FieldTyper returns the type of a field
type FieldTyper interface {
	Name() string
	Type() string
}

// FieldDefSet represents a set of field definitions
type FieldDefSet map[string]fieldType

// Add adds a FieldDef to a FieldDefSet
func (fds FieldDefSet) Add(fieldDef FieldDef) error {
	if fds[fieldDef.name] != "" {
		return fmt.Errorf("Error adding duplicate FieldDef (%v) to set", fieldDef)
	}
	fds[fieldDef.name] = fieldDef.fieldType
	return nil
}

// UnmarshalJSON unmarshals JSON to a FieldDefSet
func (fds FieldDefSet) UnmarshalJSON(b []byte) error {
	fieldDefs := make(map[string]fieldType)
	err := json.Unmarshal(b, &fieldDefs)
	if err != nil {
		return err
	}
	for key, value := range fieldDefs {
		err = isValidFieldType(value)
		if err != nil {
			return err
		}
		fd := FieldDef{
			name:      key,
			fieldType: value,
		}
		err = fds.Add(fd)
		if err != nil {
			return err
		}
	}
	return nil
}
