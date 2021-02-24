package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// SocietyPersister defines methods needed to persist societies
type SocietyPersister interface {
	SelectSocietySummariesByID(ctx context.Context, ids []uint32) ([]SocietySummary, error)
	SelectSocietySummary(ctx context.Context, id uint32) (*SocietySummary, error)
	SelectSociety(ctx context.Context) (*Society, error)
	InsertSociety(ctx context.Context, in SocietyIn) (*Society, error)
	UpdateSociety(ctx context.Context, in Society) (*Society, error)
	DeleteSociety(ctx context.Context) error
}

// SocietyBody is the JSON part of the Society object
type SocietyBody struct {
	Name         string                 `json:"name" validate:"required"`
	SecretKey    string                 `json:"secretKey" validate:"required"`
	LoginURL     string                 `json:"loginURL"`
	PostMetadata []SettingsPostMetadata `json:"postMetadata"`
}

type SettingsPostMetadata struct {
	Name    string `json:"name"  dynamodbav:"altSort"`
	Type    string `json:"type" validate:"eq=string|eq=number|eq=date|eq=boolean"`
	Tooltip string `json:"tooltip"`
}

// Value makes SocietyBody implement the driver.Valuer interface.
func (cb SocietyBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes SocietyBody implement the sql.Scanner interface.
func (cb *SocietyBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// SocietyIn is the payload to create or update a Society
type SocietyIn struct {
	SocietyBody
}

// NewSocietyIn constructs a SocietyIn
func NewSocietyIn(name, secretKey, loginURL string) SocietyIn {
	return SocietyIn{
		SocietyBody: SocietyBody{
			Name:      name,
			SecretKey: secretKey,
			LoginURL:  loginURL,
		},
	}
}

// Society represents a group of users and data
type Society struct {
	ID uint32 `json:"id,omitempty" example:"999" validate:"required,omitempty"`
	//Type string `json:"-" dynamodbav:"sk"`
	SocietyIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewSociety constructs a Society from an id and a SocietyIn
func NewSociety(id uint32, in SocietyIn) Society {
	now := time.Now()
	return Society{
		ID:             id,
		SocietyIn:      in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
}

// SocietySummary represents public info about a society
type SocietySummary struct {
	ID   uint32 `json:"id,omitempty" example:"999" validate:"required,omitempty"`
	Name string `json:"name" validate:"required"`
}
