package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// fieldType represents the type of an attribute value
type fieldType string // `enums:"Int,String,Image,Location,Time"`

const (
	// IntType indicates a field of type int
	IntType fieldType = "Int"
	// StringType indicates a field of type string
	StringType fieldType = "String"
	// ImageType indicates a field of type image.Image
	ImageType fieldType = "Image"
	// LocationType indicates a field of type Location
	LocationType fieldType = "Location"
	// TimeType indicates a field of type time.Time
	TimeType fieldType = "Time"
)

// fieldTypes returns a map of the valid field types
func fieldTypes() map[fieldType]bool {
	return map[fieldType]bool{IntType: true, StringType: true, ImageType: true, LocationType: true, TimeType: true}
}

// isValidFieldType returns an error if a fieldType isn't valid
func isValidFieldType(ft fieldType) error {
	if fieldTypes()[ft] {
		return nil
	}
	return errors.New("'" + string(ft) + "' is not a valid fieldType")
}

// func makeFieldType(s string) (fieldType, error) {
// 	ft := fieldType(s)
// 	err := isValidFieldType(ft)
// 	if err != nil {
// 		return fieldType(""), err
// 	}
// 	return ft, nil
// }

// FieldDef defines a name and type of a field
type FieldDef struct {
	Name       string    `json:"name"`
	Type       fieldType `json:"type,omitempty" swaggertype:"string" enums:"Int,String,Image,Location,Time"`
	CSVHeading string    `json:"csv_heading,omitempty"`
}

// NewFieldDef constructs a FieldDef
func NewFieldDef(name string, ft fieldType, csvHeading string) (FieldDef, error) {
	err := isValidFieldType(ft)
	if err != nil {
		return FieldDef{}, err
	}
	return FieldDef{Name: name, Type: ft, CSVHeading: csvHeading}, nil
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

// FieldDefSet represents a set of field definitions which must have both
// unigue names and unique CSV header names.
type FieldDefSet []FieldDef

// NewFieldDefSet constructs a FieldDefSet
func NewFieldDefSet() FieldDefSet {
	return FieldDefSet(make([]FieldDef, 0))
}

// Add adds a FieldDef to a FieldDefSet
func (fds *FieldDefSet) Add(fd FieldDef) bool {
	if fds.Contains(fd) {
		return false
	}
	*fds = append(*fds, fd)
	return true
}

// Contains indicates whether a FieldDefSet contains a FieldDef
func (fds *FieldDefSet) Contains(fd FieldDef) bool {
	for _, f := range *fds {
		if f.Name == fd.Name || f.CSVHeading == fd.CSVHeading {
			return true
		}
	}
	return false
}

// // MarshalJSON marshals a FieldDefSet to JSON
// func (fds FieldDefSet) MarshalJSON() ([]byte, error) {
// 	fieldDefs := make([]FieldDef, 0, len(fds.byName))
// 	for _, fd := range fds.byName {
// 		fieldDefs = append(fieldDefs, *fd)
// 	}
// 	log.Printf("fds.MarshalJSON, fieldDefs: %#v", fieldDefs)
// 	return json.Marshal(fieldDefs)
// }

// UnmarshalJSON unmarshals JSON to a FieldDefSet
func (fds *FieldDefSet) UnmarshalJSON(b []byte) error {
	if fds == nil {
		*fds = NewFieldDefSet()
	}
	var fieldDefs []FieldDef
	err := json.Unmarshal(b, &fieldDefs)
	if err != nil {
		return err
	}
	for _, value := range fieldDefs {
		err = isValidFieldType(value.Type)
		if err != nil {
			return err
		}
		if !fds.Add(value) {
			return fmt.Errorf("Attempt to add duplicate FieldDef: %#v", value)
		}
	}
	return nil
}
