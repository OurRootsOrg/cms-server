package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// PlacePersister defines methods needed to persist places
type PlacePersister interface {
	SelectPlaceMetadata(ctx context.Context) (*PlaceMetadata, error)
	SelectPlace(ctx context.Context, id uint32) (*Place, error)
	SelectPlacesByID(ctx context.Context, ids []uint32) ([]Place, error)
	SelectPlaceWord(ctx context.Context, word string) (*PlaceWord, error)
	SelectPlaceWordsByWord(ctx context.Context, words []string) ([]PlaceWord, error)
}

// Place holds information about a place
type Place struct {
	ID               uint32    `json:"id" dynamodbav:"pk"`
	Name             string    `json:"name"`
	AltNames         string    `json:"altNames"`
	Types            string    `json:"types"`
	LocatedInID      int       `json:"locatedInId"`
	AlsoLocatedInIDs string    `json:"alsoLocatedInIds"`
	Level            int       `json:"level"`
	CountryID        int       `json:"countryId"`
	Latitude         float32   `json:"latitude"`
	Longitude        float32   `json:"longitude"`
	Count            int       `json:"count"`
	InsertTime       time.Time `json:"insert_time,omitempty"`
	LastUpdateTime   time.Time `json:"last_update_time,omitempty"`
}

// PlaceWord holds the IDs of all places that have that word in their name or alt name
type PlaceWord struct {
	Word           string    `json:"word" dynamodbav:"pk"`
	IDs            string    `json:"ids"`
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// PlaceMetadataBody is the JSON body of a PlaceMetadata object
type PlaceMetadataBody struct {
	Abbreviations             map[string]string `json:"abbreviations"`
	TypeWords                 []string          `json:"typeWords"`
	NoiseWords                []string          `json:"noiseWords"`
	LargeCountries            []int             `json:"largeCountries"`
	MediumCountries           []int             `json:"mediumCountries"`
	LargeCountryLevelWeights  []int             `json:"largeCountryLevelWeights"`
	MediumCountryLevelWeights []int             `json:"mediumCountryLevelWeights"`
	SmallCountryLevelWeights  []int             `json:"smallCountryLevelWeights"`
	PrimaryMatchWeight        int               `json:"primaryMatchWeight"`
}

// Value makes PlaceMetadataBody implement the driver.Valuer interface.
func (cb PlaceMetadataBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes PlaceMetadataBody implement the sql.Scanner interface.
func (cb *PlaceMetadataBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// PlaceMetadataIn is the payload to create or update a PlaceMetadata object
type PlaceMetadataIn struct {
	PlaceMetadataBody
}

// PlaceMetadata represents global placeMetadata
type PlaceMetadata struct {
	ID int    `json:"-" dynamodbav:"pk"`
	Sk string `json:"-" dynamodbav:"sk"`
	PlaceMetadataIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewPlaceMetadataIn constructs a PlaceMetadataIn
func NewPlaceMetadataIn(abbreviations map[string]string, typeWords, noiseWords []string, largeCountries, mediumCountries []int, largeCountryLevelWeights, mediumCountryLevelWeights, smallCountryLevelWeights []int, primaryMatchWeight int) PlaceMetadataIn {
	obj := PlaceMetadataIn{
		PlaceMetadataBody: PlaceMetadataBody{
			Abbreviations:             abbreviations,
			TypeWords:                 typeWords,
			LargeCountries:            largeCountries,
			MediumCountries:           mediumCountries,
			LargeCountryLevelWeights:  largeCountryLevelWeights,
			MediumCountryLevelWeights: mediumCountryLevelWeights,
			SmallCountryLevelWeights:  smallCountryLevelWeights,
			PrimaryMatchWeight:        primaryMatchWeight,
		},
	}
	return obj
}

// NewPlaceMetadata constructs a PlaceMetadata object from a PlaceMetadataIn
func NewPlaceMetadata(obj PlaceMetadataIn) PlaceMetadata {
	now := time.Now()
	c := PlaceMetadata{
		PlaceMetadataIn: obj,
		InsertTime:      now,
		LastUpdateTime:  now,
	}
	return c
}
