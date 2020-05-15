package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// UserIDFormat is the format for User IDs.
const UserIDFormat = "/users/%d"

// UserPersister defines methods needed to persist categories
type UserPersister interface {
	SelectUsers(ctx context.Context) ([]User, error)
	SelectOneUser(ctx context.Context, id string) (User, error)
	InsertUser(ctx context.Context, in UserIn) (User, error)
	UpdateUser(ctx context.Context, id string, body User) (User, error)
	DeleteUser(ctx context.Context, id string) error
}

// UserIn is the payload to create or update a category
type UserIn struct {
	UserBody
}

// UserBody is the JSON part of the User object
type UserBody struct {
	Name           string `json:"name,omitempty" validate:"required"`
	Email          string `email:"email,omitempty" validate:"required,email"`
	EmailConfirmed bool   `email:"email_confirmed,omitempty"`
}

// NewUserIn constructs a UserIn
func NewUserIn(name string, email string, emailConfirmed bool) (UserIn, error) {
	cb, err := newUserBody(name, email, emailConfirmed)
	if err != nil {
		return UserIn{}, err
	}
	return UserIn{UserBody: cb}, nil
}

// newUserBody constructs a UserBody
func newUserBody(name string, email string, emailConfirmed bool) (UserBody, error) {
	ub := UserBody{
		Name:           name,
		Email:          email,
		EmailConfirmed: emailConfirmed,
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

// NewUserID constructs an ID for a User from an integer id
func NewUserID(id int32) string {
	return pathPrefix + fmt.Sprintf(UserIDFormat, id)
}

// User represents a set of collections that all contain the same fields
type User struct {
	ID string `json:"id,omitempty" example:"/users/999" validate:"required,omitempty"`
	UserIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewUser constructs a User from an id and body
func NewUser(id int32, in UserIn) User {
	return User{
		ID: NewUserID(id),
		UserIn: UserIn{
			UserBody: in.UserBody,
		},
	}
}
