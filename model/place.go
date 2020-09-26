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
	SelectPlaceSettings(ctx context.Context) (*PlaceSettings, error)
	SelectPlace(ctx context.Context, id uint32) (*Place, error)
	SelectPlacesByID(ctx context.Context, ids []uint32) ([]Place, error)
	SelectPlacesByFullNamePrefix(ctx context.Context, prefix string, count int) ([]Place, error)
	SelectPlaceWord(ctx context.Context, word string) (*PlaceWord, error)
	SelectPlaceWordsByWord(ctx context.Context, words []string) ([]PlaceWord, error)
}

type StringSlice []string
type Uint32Slice []uint32

// Place holds information about a place
type Place struct {
	ID               uint32      `json:"id" dynamodbav:"pk,string"`
	Type             string      `json:"-" dynamodbav:"sk"`
	AltSort          string      `json:"-" dynamodbav:"altSort"`
	Name             string      `json:"name"`
	FullName         string      `json:"fullName" dynamodbav:"fullName"`
	AltNames         StringSlice `json:"altNames" dynamodbav:"altNames"`
	Types            StringSlice `json:"types" dynamodbav:"types"`
	LocatedInID      uint32      `json:"locatedInId"`
	AlsoLocatedInIDs Uint32Slice `json:"alsoLocatedInIds" dynamodbav:"alsoLocatedInIds"`
	Level            int         `json:"level"`
	CountryID        uint32      `json:"countryId"`
	Latitude         float32     `json:"latitude"`
	Longitude        float32     `json:"longitude"`
	Count            int         `json:"count"`
	InsertTime       time.Time   `json:"insert_time,omitempty"`
	LastUpdateTime   time.Time   `json:"last_update_time,omitempty"`
}

// PlaceWord holds the IDs of all places that have that word in their name or alt name
type PlaceWord struct {
	Pk             string      `json:"-" dynamodbav:"pk"`
	Type           string      `json:"-" dynamodbav:"sk"`
	Word           string      `json:"word" dynamodbav:"-"`
	IDs            Uint32Slice `json:"ids" dynamodbav:"ids"`
	InsertTime     time.Time   `json:"insert_time,omitempty"`
	LastUpdateTime time.Time   `json:"last_update_time,omitempty"`
}

// PlaceSettingsBody is the JSON body of a PlaceSettings object
type PlaceSettingsBody struct {
	Abbreviations             map[string]string `json:"abbreviations"`
	TypeWords                 []string          `json:"typeWords" dynamodbav:"typeWords"`
	NoiseWords                []string          `json:"noiseWords" dynamodbav:"noiseWords"`
	LargeCountries            []uint32          `json:"largeCountries" dynamodbav:"largeCountries"`
	MediumCountries           []uint32          `json:"mediumCountries" dynamodbav:"mediumCountries"`
	LargeCountryLevelWeights  []int             `json:"largeCountryLevelWeights" dynamodbav:"largeCountryLevelWeights"`
	MediumCountryLevelWeights []int             `json:"mediumCountryLevelWeights" dynamodbav:"mediumCountryLevelWeights"`
	SmallCountryLevelWeights  []int             `json:"smallCountryLevelWeights" dynamodbav:"smallCountryLevelWeights"`
	PrimaryMatchWeight        int               `json:"primaryMatchWeight"`
	USCountryID               uint32            `json:"USCountryId"`
}

// Value makes PlaceSettingsBody implement the driver.Valuer interface.
func (cb PlaceSettingsBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes PlaceSettingsBody implement the sql.Scanner interface.
func (cb *PlaceSettingsBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// Value makes StringSlice implement the driver.Valuer interface.
func (cb StringSlice) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes StringSlice implement the sql.Scanner interface.
func (cb *StringSlice) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// Value makes Uint32Slice implement the driver.Valuer interface.
func (cb Uint32Slice) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes Uint32Slice implement the sql.Scanner interface.
func (cb *Uint32Slice) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// PlaceSettingsIn is the payload to create or update a PlaceSettings object
type PlaceSettingsIn struct {
	PlaceSettingsBody
}

// PlaceSettings represents global placeSettings
type PlaceSettings struct {
	ID      int    `json:"-" dynamodbav:"-"`
	Pk      string `json:"-" dynamodbav:"pk"`
	Sk      string `json:"-" dynamodbav:"sk"`
	AltSort string `json:"-" dynamodbav:"altSort"`
	PlaceSettingsIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewPlaceSettingsIn constructs a PlaceSettingsIn
func NewPlaceSettingsIn(abbreviations map[string]string, typeWords, noiseWords []string, largeCountries, mediumCountries []uint32, largeCountryLevelWeights, mediumCountryLevelWeights, smallCountryLevelWeights []int, primaryMatchWeight int, USCountryID uint32) PlaceSettingsIn {
	obj := PlaceSettingsIn{
		PlaceSettingsBody: PlaceSettingsBody{
			Abbreviations:             abbreviations,
			TypeWords:                 typeWords,
			NoiseWords:                noiseWords,
			LargeCountries:            largeCountries,
			MediumCountries:           mediumCountries,
			LargeCountryLevelWeights:  largeCountryLevelWeights,
			MediumCountryLevelWeights: mediumCountryLevelWeights,
			SmallCountryLevelWeights:  smallCountryLevelWeights,
			PrimaryMatchWeight:        primaryMatchWeight,
			USCountryID:               USCountryID,
		},
	}
	return obj
}

// NewPlaceSettings constructs a PlaceSettings object from a PlaceSettingsIn
func NewPlaceSettings(obj PlaceSettingsIn) PlaceSettings {
	now := time.Now()
	c := PlaceSettings{
		PlaceSettingsIn: obj,
		InsertTime:      now,
		LastUpdateTime:  now,
	}
	return c
}
