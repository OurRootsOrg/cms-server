package model

import (
	"text/template"

	"github.com/google/uuid"
)

// CollectionInput is the payload to create or update a Collection
type CollectionInput struct {
	Name             string            `json:"name"`
	Location         string            `json:"location"`
	CitationTemplate template.Template `json:"citationTemplate"`
}

// Collection represents a set of related Records
type Collection struct {
	CollectionInput
	Category *Category `json:"category,omitempty"`
	ID       uuid.UUID `json:"id,omitempty"`
}
