package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// SettingsPersister defines methods needed to persist settings
type SettingsPersister interface {
	SelectSettings(ctx context.Context) (*Settings, error)
	UpsertSettings(ctx context.Context, in Settings) (*Settings, error)
}

// SettingsBody is the JSON body of a Settings object
type SettingsBody struct {
	PostMetadata []SettingsPostMetadata `json:"postMetadata"`
}
type SettingsPostMetadata struct {
	Name    string `json:"name"  dynamodbav:"altSort"`
	Type    string `json:"type" validate:"eq=string|eq=number|eq=date|eq=boolean"`
	Tooltip string `json:"tooltip"`
}

// Value makes SettingsBody implement the driver.Valuer interface.
func (cb SettingsBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes SettingsBody implement the sql.Scanner interface.
func (cb *SettingsBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// SettingsIn is the payload to create or update a Settings object
type SettingsIn struct {
	SettingsBody
}

// Settings represents global settings
type Settings struct {
	ID int    `json:"-" dynamodbav:"pk"`
	Sk string `json:"-" dynamodbav:"sk"`
	SettingsIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewSettingsIn constructs a SettingsIn
func NewSettingsIn(postMetadata []SettingsPostMetadata) SettingsIn {
	obj := SettingsIn{
		SettingsBody: SettingsBody{
			PostMetadata: postMetadata,
		},
	}
	return obj
}

// NewSettings constructs a Settings object from a SettingsIn
func NewSettings(obj SettingsIn) Settings {
	now := time.Now()
	c := Settings{
		SettingsIn:     obj,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}
