package model

import (
	"context"
	"time"
)

type NameType int

const (
	GivenType NameType = iota
	SurnameType
)

// NamePersister defines methods needed to persist places
type NamePersister interface {
	SelectNameVariants(ctx context.Context, nameType NameType, name string) (*NameVariants, error)
}

// NameVariants holds name variants
type NameVariants struct {
	Name           string      `json:"name" dynamodbav:"pk"`
	Type           string      `json:"-" dynamodbav:"sk"`
	Variants       StringSlice `json:"variants"`
	InsertTime     time.Time   `json:"insert_time,omitempty"`
	LastUpdateTime time.Time   `json:"last_update_time,omitempty"`
}
