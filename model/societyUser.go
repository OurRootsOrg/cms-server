package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type AuthLevel int

const (
	AuthGuest = iota
	AuthReader
	AuthContributor
	AuthEditor
	AuthAdmin
)

func (l AuthLevel) String() string {
	return [...]string{"Guest", "Reader", "Contributor", "Editor", "Admin"}[l]
}

// SocietyUserPersister defines methods needed to persist SocietyUsers
type SocietyUserPersister interface {
	SelectSocietyUsers(ctx context.Context) ([]SocietyUser, error)
	SelectAllSocietyUsersByUser(ctx context.Context, userID uint32) ([]SocietyUser, error)
	SelectOneSocietyUser(ctx context.Context, id uint32) (*SocietyUser, error)
	SelectOneSocietyUserByUser(ctx context.Context, userID uint32) (*SocietyUser, error)
	InsertSocietyUser(ctx context.Context, in SocietyUserIn) (*SocietyUser, error)
	UpdateSocietyUser(ctx context.Context, id uint32, in SocietyUser) (*SocietyUser, error) // can't update userID or societyID
	DeleteSocietyUser(ctx context.Context, id uint32) error
}

// SocietyUserBody is the JSON part of the SocietyUser object
type SocietyUserBody struct {
	UserName string    `json:"name"` // override userName from user record
	Level    AuthLevel `json:"level" validate:"required"`
}

// Value makes SocietyUserBody implement the driver.Valuer interface.
func (cb SocietyUserBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes SocietyUserBody implement the sql.Scanner interface.
func (cb *SocietyUserBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// SocietyUserIn is the payload to create or update a SocietyUser
type SocietyUserIn struct {
	SocietyUserBody
	UserID    uint32 `json:"userId" validate:"required"`
	SocietyID uint32 `json:"societyId" validate:"required"`
}

// NewSocietyUserIn constructs a SocietyUserIn
func NewSocietyUserIn(userID uint32, level AuthLevel) SocietyUserIn {
	return SocietyUserIn{
		SocietyUserBody: SocietyUserBody{
			Level: level,
		},
		UserID: userID,
	}
}

// SocietyUser represents an authorization level between a user and a society
type SocietyUser struct {
	ID uint32 `json:"id,omitempty" example:"999" validate:"required,omitempty"`
	//	Type string `json:"-" dynamodbav:"sk"`
	SocietyUserIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewSocietyUser constructs a SocietyUser from an id and a SocietyUserIn
func NewSocietyUser(id uint32, in SocietyUserIn) SocietyUser {
	now := time.Now()
	return SocietyUser{
		ID:             id,
		SocietyUserIn:  in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
}
