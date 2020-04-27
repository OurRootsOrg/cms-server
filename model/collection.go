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
	Name             string             `json:"name,omitempty"`
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
// func NewCollectionBody(name string, category *Category, location string, template *template.Template) CollectionBody {
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
	Category CategoryRef `json:"category"`
}

// CollectionRef is a reference to a Category
type CollectionRef struct {
	ID   string `json:"id,omitempty" example:"/collections/999"`
	Type string `json:"type,omitempty" example:"collection"`
}

// NewCollectionRef constructs a CollectionRef from an id
func NewCollectionRef(id int32) CollectionRef {
	return CollectionRef{
		ID:   fmt.Sprintf(pathPrefix+CollectionIDFormat, id),
		Type: "collection",
	}
}

// Collection represents a set of related Records
type Collection struct {
	CollectionRef
	CollectionBody
	Category       CategoryRef `json:"category"`
	InsertTime     time.Time   `json:"insert_time,omitempty"`
	LastUpdateTime time.Time   `json:"last_update_time,omitempty"`
}

// NewCollection constructs a Collection from a CollectionIn
func NewCollection(id int32, ci CollectionIn) Collection {
	c := Collection{
		CollectionRef:  NewCollectionRef(id),
		CollectionBody: ci.CollectionBody,
		Category:       ci.Category,
		InsertTime:     time.Now(),
		LastUpdateTime: time.Now(),
	}
	return c
}
