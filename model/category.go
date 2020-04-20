package model

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

// CategoryInput is the payload to create or update a category
type CategoryInput struct {
	Name       string      `json:"name,omitempty"`
	CSVHeading string      `json:"csv_heading,omitempty"`
	FieldDefs  FieldDefSet `json:"field_defs,omitempty"` // swaggertype:"map[string]string" example:"{\"int_field\":\"Int\", \"string_field\":\"String\"}"
	//IndexMapping
}

// NewCategoryInput constructs a CategoryInput
func NewCategoryInput(name string, fieldDefs ...FieldDef) (CategoryInput, error) {
	ci := CategoryInput{Name: name, FieldDefs: NewFieldDefSet()}
	for _, fd := range fieldDefs {
		if !ci.FieldDefs.Add(fd) {
			return CategoryInput{}, fmt.Errorf("Attempt to add duplicate FieldDef: %#v", fd)
		}
	}
	return ci, nil
}

// Category represents a set of collections that all contain the same fields
type Category struct {
	CategoryInput
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

// NewCategory constructs a Category from a CategoryInput
func NewCategory(ci CategoryInput) Category {
	id := uuid.New()
	u, err := url.Parse("/categories/" + id.String())
	if err != nil {
		// Should never happen
		panic(err)
	}
	c := Category{
		CategoryInput: ci,
		ID:            u.String(),
		Type:          "category",
	}
	return c
}
