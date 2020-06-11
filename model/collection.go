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

// CollectionIDFormat is the format for Collection IDs
const CollectionIDFormat = "/collections/%d"

// CollectionPersister defines methods needed to persist categories
type CollectionPersister interface {
	SelectCollections(ctx context.Context) ([]Collection, error)
	SelectManyCollections(ctx context.Context, ids []string) ([]Collection, error)
	SelectOneCollection(ctx context.Context, id string) (Collection, error)
	InsertCollection(ctx context.Context, in CollectionIn) (Collection, error)
	UpdateCollection(ctx context.Context, id string, in Collection) (Collection, error)
	DeleteCollection(ctx context.Context, id string) error
}

// CollectionBody is the JSON body of a Collection
type CollectionBody struct {
	Name             string             `json:"name,omitempty" validate:"required,omitempty"`
	Location         string             `json:"location,omitempty"`
	Fields           []CollectionField  `json:"fields"`
	CitationTemplate *template.Template `json:"citation_template,omitempty"`
}
type CollectionField struct {
	Header     string `json:"header"`
	Name       string `json:"name"`
	Required   bool   `json:"required,omitempty"`
	Regex      string `json:"regex,omitempty"`
	RegexError string `json:"regexError,omitempty"`
}

// Value makes CollectionBody implement the driver.Valuer interface.
func (cb CollectionBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes CollectionBody implement the sql.Scanner interface.
func (cb *CollectionBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// CollectionIn is the payload to create or update a Collection
type CollectionIn struct {
	CollectionBody
	Category string `json:"category,omitempty" example:"/categories/999" validate:"required"`
}

// Collection represents a set of related Records
type Collection struct {
	ID string `json:"id,omitempty" example:"/collections/999" validate:"required"`
	CollectionIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewCollectionIn constructs a CollectionIn
func NewCollectionIn(name string, categoryID string) CollectionIn {
	ci := CollectionIn{
		CollectionBody: CollectionBody{
			Name: name,
		},
		Category: categoryID,
	}
	return ci
}

// NewCollection constructs a Collection from a CollectionIn
func NewCollection(id int32, ci CollectionIn) Collection {
	now := time.Now()
	c := Collection{
		ID: MakeCollectionID(id),
		CollectionIn: CollectionIn{
			CollectionBody: ci.CollectionBody,
			Category:       ci.Category,
		},
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}

// MakeCollectionID builds a Collection ID string from an integer ID
func MakeCollectionID(id int32) string {
	return pathPrefix + fmt.Sprintf(CollectionIDFormat, id)
}
