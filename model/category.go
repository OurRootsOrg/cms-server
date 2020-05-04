package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Initialize sets up app-specific values for the package
func Initialize(p string) {
	pathPrefix = p
}

var pathPrefix string

// CategoryIDFormat is the format for Category IDs.
const CategoryIDFormat = "/categories/%d"

// CategoryPersister defines methods needed to persist categories
type CategoryPersister interface {
	SelectCategories() ([]Category, error)
	SelectOneCategory(id string) (Category, error)
	InsertCategory(in CategoryIn) (Category, error)
	UpdateCategory(id string, body CategoryIn) (Category, error)
	DeleteCategory(id string) error
}

// CategoryIn is the payload to create or update a category
type CategoryIn struct {
	CategoryBody
}

// CategoryBody is the JSON part of the Category object
type CategoryBody struct {
	Name      string      `json:"name,omitempty" validate:"required"`
	FieldDefs FieldDefSet `json:"field_defs,omitempty"` //   example:"{\"int_field\":\"Int\", \"string_field\":\"String\"}"
	//IndexMapping
}

// NewCategoryIn constructs a CategoryIn
func NewCategoryIn(name string, fieldDefs ...FieldDef) (CategoryIn, error) {
	cb, err := newCategoryBody(name, fieldDefs...)
	if err != nil {
		return CategoryIn{}, err
	}
	return CategoryIn{CategoryBody: cb}, nil
}

// newCategoryBody constructs a CategoryBody
func newCategoryBody(name string, fieldDefs ...FieldDef) (CategoryBody, error) {
	cb := CategoryBody{Name: name}
	for _, fd := range fieldDefs {
		if !cb.FieldDefs.Add(fd) {
			return CategoryBody{}, fmt.Errorf("Attempt to add duplicate FieldDef: %#v", fd)
		}
	}
	return cb, nil
}

// Value makes CategoryBody implement the driver.Valuer interface.
func (cb CategoryBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes CategoryBody implement the sql.Scanner interface.
func (cb *CategoryBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// CategoryRef is a reference to a Category
type CategoryRef struct {
	ID   string `json:"id,omitempty" example:"/categories/999" validate:"required,omitempty"`
	Type string `json:"type,omitempty" example:"category" validate:"required,omitempty"`
}

// NewCategoryRef constructs a CategoryRef from an id
func NewCategoryRef(id int32) CategoryRef {
	return CategoryRef{
		ID:   pathPrefix + fmt.Sprintf(CategoryIDFormat, id),
		Type: "category",
	}
}

// Value makes CategoryRef implement the driver.Valuer interface.
func (cr CategoryRef) Value() (driver.Value, error) {
	var catID int64
	_, err := fmt.Sscanf(cr.ID, pathPrefix+CategoryIDFormat, &catID)
	return catID, err
}

// Scan makes CategoryRef implement the sql.Scanner interface.
func (cr *CategoryRef) Scan(value interface{}) error {
	catID, ok := value.(int64)
	if !ok {
		return fmt.Errorf("type assertion to int64 failed: %v", value)
	}
	cr.ID = pathPrefix + fmt.Sprintf(CategoryIDFormat, catID)
	cr.Type = "category"
	return nil
}

// Category represents a set of collections that all contain the same fields
type Category struct {
	CategoryRef
	CategoryBody
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewCategory constructs a Category from an id and body
func NewCategory(id int32, in CategoryIn) Category {
	return Category{
		CategoryRef:  NewCategoryRef(id),
		CategoryBody: in.CategoryBody,
	}
}
