package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"time"
)

// CollectionIDFormat is the format for Collection IDs
const CollectionIDFormat = "/collections/%d"

var CollectionName = "collection"

// CollectionPersister defines methods needed to persist categories
type CollectionPersister interface {
	SelectCollections() ([]Collection, error)
	SelectOneCollection(id string) (Collection, error)
	InsertCollection(in CollectionIn) (Collection, error)
	UpdateCollection(id string, in CollectionIn) (Collection, error)
	DeleteCollection(id string) error
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

// NewCollectionBody builds a CollectionBody
// func NewCollectionBody(name string, category Category, location string, template template.Template) CollectionBody {
// 	// TODO: validation
// 	return CollectionBody{
// 		Name:             name,
// 		Location:         location,
// 		CitationTemplate: template,
// 	}
// }

// CollectionIn is the payload to create or update a Collection
type CollectionIn struct {
	CollectionBody
	Category CategoryRef `json:"category,omitempty" validate:"required,omitempty"`
}

// CollectionRef is a reference to a Category
type CollectionRef struct {
	ID   string `json:"id,omitempty" example:"/collections/999" validate:"required"`
	Type string `json:"type,omitempty" example:"collection" validate:"required"`
}

// NewCollectionRef constructs a CollectionRef from an id
func NewCollectionRef(id int32) CollectionRef {
	cid := fmt.Sprintf(pathPrefix+CollectionIDFormat, id)
	return CollectionRef{
		ID:   cid,
		Type: CollectionName,
	}
}

// Collection represents a set of related Records
type Collection struct {
	CollectionRef
	CollectionBody
	Category       CategoryRef `json:"category" validate:"required"`
	InsertTime     time.Time   `json:"insert_time,omitempty"`
	LastUpdateTime time.Time   `json:"last_update_time,omitempty"`
}

// NewCollection constructs a Collection from a CollectionIn
func NewCollection(id int32, ci CollectionIn) Collection {
	now := time.Now()
	c := Collection{
		CollectionRef:  NewCollectionRef(id),
		CollectionBody: ci.CollectionBody,
		Category:       ci.Category,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}
