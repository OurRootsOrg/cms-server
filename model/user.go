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
	Name         string `json:"name,omitempty" validate:"required"`
	Email        string `email:"email,omitempty" validate:"required,email"`
	passwordHash string
}

// NewUserIn constructs a UserIn
func NewUserIn(name string, email string, passwordHash string) (UserIn, error) {
	cb, err := newUserBody(name, email, passwordHash)
	if err != nil {
		return UserIn{}, err
	}
	return UserIn{UserBody: cb}, nil
}

// newUserBody constructs a UserBody
func newUserBody(name string, email string, passwordHash string) (UserBody, error) {
	ub := UserBody{
		Name:         name,
		Email:        email,
		passwordHash: passwordHash,
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

// UserRef is a reference to a User
type UserRef struct {
	ID   string `json:"id,omitempty" example:"/users/999" validate:"required,omitempty"`
	Type string `json:"type,omitempty" example:"user" validate:"required,omitempty"`
}

// NewUserRef constructs a UserRef from an id
func NewUserRef(id int32) UserRef {
	return UserRef{
		ID:   pathPrefix + fmt.Sprintf(UserIDFormat, id),
		Type: "category",
	}
}

// Value makes UserRef implement the driver.Valuer interface.
func (cr UserRef) Value() (driver.Value, error) {
	var catID int64
	_, err := fmt.Sscanf(cr.ID, pathPrefix+UserIDFormat, &catID)
	return catID, err
}

// Scan makes UserRef implement the sql.Scanner interface.
func (cr *UserRef) Scan(value interface{}) error {
	catID, ok := value.(int64)
	if !ok {
		return fmt.Errorf("type assertion to int64 failed: %v", value)
	}
	cr.ID = pathPrefix + fmt.Sprintf(UserIDFormat, catID)
	cr.Type = "category"
	return nil
}

// User represents a set of collections that all contain the same fields
type User struct {
	UserRef
	UserIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewUser constructs a User from an id and body
func NewUser(id int32, in UserIn) User {
	return User{
		UserRef: NewUserRef(id),
		UserIn: UserIn{
			UserBody: in.UserBody,
		},
	}
}
