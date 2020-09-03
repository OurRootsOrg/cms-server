package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// UserPersister defines methods needed to persist categories
type UserPersister interface {
	RetrieveUser(ctx context.Context, in UserIn) (*User, error)
	// SelectUsers(ctx context.Context) ([]User, error)
	// SelectOneUser(ctx context.Context, id string) (User, error)
	// InsertUser(ctx context.Context, in UserIn) (User, error)
	// UpdateUser(ctx context.Context, id string, body User) (User, error)
	// DeleteUser(ctx context.Context, id string) error
}

// UserIn is the payload to create or update a category
type UserIn struct {
	UserBody
}

// UserBody is the JSON part of the User object
type UserBody struct {
	Name           string `json:"name,omitempty" validate:"required"`
	Email          string `json:"email,omitempty" validate:"required,email"`
	EmailConfirmed bool   `json:"email_confirmed,omitempty"`
	Issuer         string `json:"iss" validate:"required,url" dynamodbav:"-"`
	Subject        string `json:"sub" validate:"required" dynamodbav:"-"`
	Enabled        bool   `json:"enabled"`
}

// NewUserIn constructs a UserIn
func NewUserIn(name string, email string, emailConfirmed bool, issuer string, subject string) (UserIn, error) {
	cb, err := newUserBody(name, email, emailConfirmed, issuer, subject, true)
	if err != nil {
		return UserIn{}, err
	}
	return UserIn{UserBody: cb}, nil
}

// newUserBody constructs a UserBody
func newUserBody(name string, email string, emailConfirmed bool, issuer string, subject string, enabled bool) (UserBody, error) {
	ub := UserBody{
		Name:           name,
		Email:          email,
		EmailConfirmed: emailConfirmed,
		Issuer:         issuer,
		Subject:        subject,
		Enabled:        enabled,
	}
	return ub, nil
}

// Value makes UserBody implement the driver.Valuer interface.
func (cb UserBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes UserBody implement the sql.Scanner interface.
func (cb *UserBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// User represents a set of collections that all contain the same fields
type User struct {
	ID      uint32 `json:"id,omitempty" example:"999" validate:"required,omitempty" dynamodbav:"pk,string"`
	Type    string `json:"-" dynamodbav:"sk"`
	SortKey string `json:"-" dynamodbav:"altSort"`
	UserBody
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewUser constructs a User from an id and body
func NewUser(id uint32, in UserIn) User {
	return User{
		ID:       id,
		UserBody: in.UserBody,
	}
}
