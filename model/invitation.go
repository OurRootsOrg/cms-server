package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// InvitationPersister defines methods needed to persist Invitations
type InvitationPersister interface {
	SelectInvitationByCode(ctx context.Context, code string) (*Invitation, error)
	SelectInvitations(ctx context.Context) ([]Invitation, error)
	SelectOneInvitation(ctx context.Context, id uint32) (*Invitation, error)
	InsertInvitation(ctx context.Context, in InvitationIn) (*Invitation, error)
	DeleteInvitation(ctx context.Context, id uint32) error
}

// InvitationBody is the JSON part of the Invitation object
type InvitationBody struct {
	Level AuthLevel `json:"level" validate:"required"`
	Name  string    `json:"name" validate:"required"`
}

// Value makes InvitationBody implement the driver.Valuer interface.
func (cb InvitationBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes InvitationBody implement the sql.Scanner interface.
func (cb *InvitationBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// InvitationIn is the payload to create or update a Invitation
type InvitationIn struct {
	InvitationBody
	Code      string `json:"code" validate:"required"`
	SocietyID uint32 `json:"societyId" validate:"required"`
}

// NewInvitationIn constructs an InvitationIn
func NewInvitationIn(name, code string, level AuthLevel, societyID uint32) InvitationIn {
	return InvitationIn{
		InvitationBody: InvitationBody{
			Level: level,
			Name:  name,
		},
		Code:      code,
		SocietyID: societyID,
	}
}

// Invitation represents an invitation to a society
type Invitation struct {
	ID uint32 `json:"id,omitempty" example:"999" validate:"required,omitempty"`
	//Type string `json:"-" dynamodbav:"sk"`
	InvitationIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewInvitation constructs a Invitation from an id and body
func NewInvitation(id uint32, in InvitationIn) Invitation {
	now := time.Now()
	return Invitation{
		ID:             id,
		InvitationIn:   in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
}
