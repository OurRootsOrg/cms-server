package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RecordIDFormat is the format for Record IDs
const RecordIDFormat = "/records/%d"

// RecordPersister defines methods needed to persist records
type RecordPersister interface {
	SelectRecordsForPost(ctx context.Context, postID string) ([]Record, error)
	SelectOneRecord(ctx context.Context, id string) (Record, error)
	InsertRecord(ctx context.Context, in RecordIn) (Record, error)
	UpdateRecord(ctx context.Context, id string, in Record) (Record, error)
	DeleteRecord(ctx context.Context, id string) error
	DeleteRecordsForPost(ctx context.Context, postID string) error
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
	Post string `json:"post,omitempty" example:"/posts/999" validate:"required"`
}

// Record represents a set of related Records
type Record struct {
	ID string `json:"id,omitempty" example:"/records/999" validate:"required"`
	RecordIn
	IxHash         string    `json:"ix_hash,omitempty"`
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewRecordIn constructs a RecordIn
func NewRecordIn(data map[string]string, postID string) RecordIn {
	ri := RecordIn{
		RecordBody: RecordBody{
			Data: data,
		},
		Post: postID,
	}
	return ri
}

// NewRecord constructs a Record from a RecordIn
func NewRecord(id int32, ci RecordIn) Record {
	now := time.Now()
	c := Record{
		ID: MakeRecordID(id),
		RecordIn: RecordIn{
			RecordBody: ci.RecordBody,
			Post:       ci.Post,
		},
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}

// MakeRecordID builds a Record ID string from an integer ID
func MakeRecordID(id int32) string {
	return pathPrefix + fmt.Sprintf(RecordIDFormat, id)
}
