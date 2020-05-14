package model

import (
	"context"
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
	SelectCategories(ctx context.Context) ([]Category, error)
	SelectOneCategory(ctx context.Context, id string) (Category, error)
	InsertCategory(ctx context.Context, in CategoryIn) (Category, error)
	UpdateCategory(ctx context.Context, id string, body Category) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
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

// Category represents a set of collections that all contain the same fields
type Category struct {
	ID string `json:"id,omitempty" example:"/categories/999" validate:"required,omitempty"`
	CategoryIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewCategory constructs a Category from an id and body
func NewCategory(id int32, in CategoryIn) Category {
	return Category{
		ID: MakeCategoryID(id),
		CategoryIn: CategoryIn{
			CategoryBody: in.CategoryBody,
		},
	}
}

// MakeCategoryID builds a Category ID string from an integer ID
func MakeCategoryID(id int32) string {
	return pathPrefix + fmt.Sprintf(CategoryIDFormat, id)
}
