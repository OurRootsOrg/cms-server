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
	SelectSettings(ctx context.Context) (Settings, error)
	UpsertSettings(ctx context.Context, in Settings) (Settings, error)
}

// SettingsBody is the JSON body of a Settings object
type SettingsBody struct {
	PostFields []SettingsPostField `json:"postFields"`
}
type SettingsPostField struct {
	Name string `json:"name"`
	Type string `json:"type" validate:"eq=string|eq=number|eq=date|eq=boolean|eq=rating"`
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
	SettingsIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewSettingsIn constructs a SettingsIn
func NewSettingsIn(postFields []SettingsPostField) SettingsIn {
	obj := SettingsIn{
		SettingsBody: SettingsBody{
			PostFields: postFields,
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
