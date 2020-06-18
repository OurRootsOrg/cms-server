package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"time"
)

// Initialize sets up app-specific values for the package
func Initialize(p string) {
	pathPrefix = p
}

var pathPrefix string

// CategoryPersister defines methods needed to persist categories
type CategoryPersister interface {
	SelectCategories(ctx context.Context) ([]Category, error)
	SelectOneCategory(ctx context.Context, id uint32) (Category, error)
	InsertCategory(ctx context.Context, in CategoryIn) (Category, error)
	UpdateCategory(ctx context.Context, id uint32, body Category) (Category, error)
	DeleteCategory(ctx context.Context, id uint32) error
}

// CategoryIn is the payload to create or update a category
type CategoryIn struct {
	CategoryBody
}

// CategoryBody is the JSON part of the Category object
type CategoryBody struct {
	Name                      string             `json:"name,omitempty" validate:"required"`
	FieldDefs                 FieldDefSet        `json:"field_defs,omitempty"` // example:"{\"int_field\":\"Int\", \"string_field\":\"String\"}"
	SearchResultsPageTemplate *template.Template `json:"search_results_page_template,omitempty"`
	DetailsPageTemplate       *template.Template `json:"details_page_template,omitempty"`
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
	ID uint32 `json:"id,omitempty" example:"999" validate:"required,omitempty"`
	CategoryBody
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewCategory constructs a Category from an id and body
func NewCategory(id uint32, in CategoryIn) Category {
	return Category{
		ID:           id,
		CategoryBody: in.CategoryBody,
	}
}
