package model

import (
	"net/url"
	"text/template"

	"github.com/google/uuid"
)

// CollectionInput is the payload to create or update a Collection
type CollectionInput struct {
	Name             string             `json:"name,omitempty"`
	Location         *string            `json:"location,omitempty"`
	Category         *Category          `json:"category,omitempty"`
	CitationTemplate *template.Template `json:"citation_template,omitempty"`
}

// Collection represents a set of related Records
type Collection struct {
	CollectionInput
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

// NewCollection constructs a Collection from a CollectionInput
func NewCollection(ci CollectionInput) Collection {
	id := uuid.New()
	u, err := url.Parse("/collections/" + id.String())
	if err != nil {
		panic(err)
	}
	c := Collection{
		CollectionInput: ci,
		ID:              u.String(),
		Type:            "collection",
	}
	return c
}
