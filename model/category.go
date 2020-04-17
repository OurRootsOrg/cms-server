package model

import (
	"github.com/google/uuid"
)

// CategoryInput is the payload to create or update a category
type CategoryInput struct {
	Name       string `json:"name"`
	CSVHeading string `json:"csvHeading"`
	//IndexMapping
	FieldDefs FieldDefSet `json:"fieldDefs"`
}

// NewCategoryInput constructs a CategoryInput
func NewCategoryInput(name string, fieldDefs ...FieldDef) (CategoryInput, error) {
	ci := CategoryInput{Name: name, FieldDefs: make(FieldDefSet)}
	for _, fd := range fieldDefs {
		err := ci.FieldDefs.Add(fd)
		if err != nil {
			return CategoryInput{}, err
		}
	}
	return ci, nil
}

// Category represents a set of collections that all contain the same fields
type Category struct {
	CategoryInput
	ID uuid.UUID `json:"id,omitempty"`
}

// NewCategory constructs a Category from a CategoryInput
func NewCategory(ci CategoryInput) Category {
	id := uuid.New()
	c := Category{
		CategoryInput: ci,
		ID:            id,
	}
	return c
}
