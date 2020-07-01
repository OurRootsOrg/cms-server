package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// RecordPersister defines methods needed to persist records
type RecordPersister interface {
	SelectRecordsForPost(ctx context.Context, postID uint32) ([]Record, error)
	SelectRecordsByID(ctx context.Context, ids []uint32) ([]Record, error)
	SelectOneRecord(ctx context.Context, id uint32) (Record, error)
	InsertRecord(ctx context.Context, in RecordIn) (Record, error)
	UpdateRecord(ctx context.Context, id uint32, in Record) (Record, error)
	DeleteRecord(ctx context.Context, id uint32) error
	DeleteRecordsForPost(ctx context.Context, postID uint32) error
}

// RecordBody is the JSON body of a Record
type RecordBody struct {
	Data map[string]string `json:"data" validate:"required"`
}

// Value makes RecordBody implement the driver.Valuer interface.
func (cb RecordBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes RecordBody implement the sql.Scanner interface.
func (cb *RecordBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// RecordIn is the payload to create or update a Record
type RecordIn struct {
	RecordBody
	Post uint32 `json:"post" example:"999" validate:"required"`
}

// Record represents a set of related Records
type Record struct {
	ID uint32 `json:"id,omitempty" example:"999" validate:"required"`
	RecordIn
	IxHash         string    `json:"ix_hash,omitempty"`
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewRecordIn constructs a RecordIn
func NewRecordIn(data map[string]string, postID uint32) RecordIn {
	ri := RecordIn{
		RecordBody: RecordBody{
			Data: data,
		},
		Post: postID,
	}
	return ri
}

// NewRecord constructs a Record from a RecordIn
func NewRecord(id uint32, ci RecordIn) Record {
	now := time.Now()
	c := Record{
		ID: id,
		RecordIn: RecordIn{
			RecordBody: ci.RecordBody,
			Post:       ci.Post,
		},
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}
