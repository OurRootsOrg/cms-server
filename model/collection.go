package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type PrivacyLevel int

const (
	PrivacyPublic = iota
	PrivacyPrivateImages
	PrivacyPrivateDetail
	PrivacyPrivateDetailImages
	PrivacyPrivateSearch
	PrivacyPrivateSearchImages
	PrivacyPrivateSearchDetail
	PrivacyPrivateSearchDetailImages
)

func (l PrivacyLevel) String() string {
	return [...]string{"Public", "PrivateImages", "PrivateDetail", "PrivateDetailImages", "PrivateSearch",
		"PrivateSearchImages", "PrivateSearchDetail", "PrivateSearchDetailImages"}[l]
}

// CollectionPersister defines methods needed to persist categories
type CollectionPersister interface {
	SelectCollections(ctx context.Context) ([]Collection, error)
	SelectCollectionsByID(ctx context.Context, ids []uint32, enforceContextSocietyMatch bool) ([]Collection, error)
	SelectOneCollection(ctx context.Context, id uint32) (*Collection, error)
	InsertCollection(ctx context.Context, in CollectionIn) (*Collection, error)
	UpdateCollection(ctx context.Context, id uint32, in Collection) (*Collection, error)
	DeleteCollection(ctx context.Context, id uint32) error
}

// CollectionBody is the JSON body of a Collection
type CollectionBody struct {
	Name                        string              `json:"name" validate:"required" dynamodbav:"altSort"`
	Location                    string              `json:"location,omitempty"`
	Fields                      []CollectionField   `json:"fields"`
	Mappings                    []CollectionMapping `json:"mappings"`
	CitationTemplate            string              `json:"citation_template,omitempty"`
	ImagePathHeader             string              `json:"imagePathHeader,omitempty"`
	HouseholdNumberHeader       string              `json:"householdNumberHeader,omitempty"`
	HouseholdRelationshipHeader string              `json:"householdRelationshipHeader,omitempty"`
	GenderHeader                string              `json:"genderHeader,omitempty"`
	PrivacyLevel                PrivacyLevel        `json:"privacyLevel"`
}

type CollectionField struct {
	Header string `json:"header"`
}

type CollectionMapping struct {
	Header  string `json:"header"`
	DbField string `json:"dbField"`
	IxRole  string `json:"ixRole"`
	IxField string `json:"ixField"`
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
	Categories []uint32 `json:"categories" validate:"required"`
}

// Collection represents a set of related Records
type Collection struct {
	ID   uint32 `json:"id,omitempty" example:"999" validate:"required" dynamodbav:"pk,string"`
	Type string `json:"-" dynamodbav:"sk"`
	CollectionIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewCollectionIn constructs a CollectionIn
func NewCollectionIn(name string, categoryIDs []uint32) CollectionIn {
	ci := CollectionIn{
		CollectionBody: CollectionBody{
			Name: name,
		},
		Categories: categoryIDs,
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
			Categories:     ci.Categories,
		},
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}
