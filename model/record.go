package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"html/template"
	"regexp"
	"strings"
	"time"
)

// RecordPersister defines methods needed to persist records
type RecordPersister interface {
	SelectRecordsForPost(ctx context.Context, postID uint32) ([]Record, error)
	SelectRecordsByID(ctx context.Context, ids []uint32) ([]Record, error)
	SelectOneRecord(ctx context.Context, id uint32) (*Record, error)
	InsertRecord(ctx context.Context, in RecordIn) (*Record, error)
	UpdateRecord(ctx context.Context, id uint32, in Record) (*Record, error)
	DeleteRecord(ctx context.Context, id uint32) error
	DeleteRecordsForPost(ctx context.Context, postID uint32) error
	SelectRecordHouseholdsForPost(ctx context.Context, postID uint32) ([]RecordHousehold, error)
	SelectOneRecordHousehold(ctx context.Context, postID uint32, householdID string) (*RecordHousehold, error)
	InsertRecordHousehold(ctx context.Context, in RecordHouseholdIn) (*RecordHousehold, error)
	DeleteRecordHouseholdsForPost(ctx context.Context, postID uint32) error
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
	Post uint32 `json:"post" example:"999" validate:"required" dynamodbav:"-"`
}

// Record represents a set of related Records
type Record struct {
	ID   uint32 `json:"id,omitempty" example:"999" validate:"required" dynamodbav:"pk,string"`
	Type string `json:"-" dynamodbav:"sk"`
	RecordIn
	IxHash         string    `json:"ix_hash,omitempty"`
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// RecordHouseholdIn is the payload to create a Record Household
type RecordHouseholdIn struct {
	Post      uint32      `json:"post" example:"999" validate:"required"`
	Household string      `json:"household" example:"999" validate:"required"`
	Records   Uint32Slice `json:"records" example:"[1,2,3]" validate:"required"`
}

// RecordHousehold holds the record IDs of all records in this household
type RecordHousehold struct {
	RecordHouseholdIn
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

var mustacheRE = regexp.MustCompile(`{{\s*([^} ]+(\s+[^} ]+)*)\s*}}`)

// Accept a citation in simple mustache syntax {{ var name can have spaces }}
// and convert it to go template syntax before applying it to the record data
func (record Record) GetCitation(tpl string) string {
	var citation strings.Builder
	if tpl != "" {
		tpl = mustacheRE.ReplaceAllString(tpl, "{{index . \"$1\"}}")
		tmpl, err := template.New("citation").Parse(tpl)
		if err != nil {
			citation.WriteString(err.Error())
		} else if err = tmpl.Execute(&citation, record.Data); err != nil {
			citation.WriteString(err.Error())
		}
	}
	return citation.String()
}
