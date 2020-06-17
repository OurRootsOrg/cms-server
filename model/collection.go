package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"html/template"
	"time"
)

// CollectionPersister defines methods needed to persist categories
type CollectionPersister interface {
	SelectCollections(ctx context.Context) ([]Collection, error)
	SelectCollectionsByID(ctx context.Context, ids []uint32) ([]Collection, error)
	SelectOneCollection(ctx context.Context, id uint32) (Collection, error)
	InsertCollection(ctx context.Context, in CollectionIn) (Collection, error)
	UpdateCollection(ctx context.Context, id uint32, in Collection) (Collection, error)
	DeleteCollection(ctx context.Context, id uint32) error
}

// CollectionBody is the JSON body of a Collection
type CollectionBody struct {
	Name             string             `json:"name,omitempty" validate:"required,omitempty"`
	Location         string             `json:"location,omitempty"`
	CitationTemplate *template.Template `json:"citation_template,omitempty"`
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
	Category uint32 `json:"category,omitempty" example:"999" validate:"required"`
}

// Collection represents a set of related Records
type Collection struct {
	ID uint32 `json:"id,omitempty" example:"999" validate:"required"`
	CollectionIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewCollectionIn constructs a CollectionIn
func NewCollectionIn(name string, categoryID uint32) CollectionIn {
	ci := CollectionIn{
		CollectionBody: CollectionBody{
			Name: name,
		},
		Category: categoryID,
	}
	return ci
}

// NewCollection constructs a Collection from a CollectionIn
func NewCollection(id uint32, ci CollectionIn) Collection {
	now := time.Now()
	c := Collection{
		ID: id,
		CollectionIn: CollectionIn{
			CollectionBody: ci.CollectionBody,
			Category:       ci.Category,
		},
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}
