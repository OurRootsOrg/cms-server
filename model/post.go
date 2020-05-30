package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// PostIDFormat is the format for Post IDs
const PostIDFormat = "/posts/%d"

// PostPersister defines methods needed to persist categories
type PostPersister interface {
	SelectPosts(ctx context.Context) ([]Post, error)
	SelectOnePost(ctx context.Context, id string) (Post, error)
	InsertPost(ctx context.Context, in PostIn) (Post, error)
	UpdatePost(ctx context.Context, id string, in Post) (Post, error)
	DeletePost(ctx context.Context, id string) error
}

// PostBody is the JSON body of a Post
type PostBody struct {
	Name          string `json:"name,omitempty" validate:"required,omitempty"`
	RecordsKey    string `json:"recordsKey" validate:"required"`
	RecordsStatus string `json:"recordsStatus"`
}

// Value makes PostBody implement the driver.Valuer interface.
func (cb PostBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes PostBody implement the sql.Scanner interface.
func (cb *PostBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// PostIn is the payload to create or update a Post
type PostIn struct {
	PostBody
	Collection string `json:"collection,omitempty" example:"/collections/999" validate:"required"`
}

// Post represents a set of related Records
type Post struct {
	ID string `json:"id,omitempty" example:"/posts/999" validate:"required"`
	PostIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewPostIn constructs a PostIn
func NewPostIn(name string, collectionID string) PostIn {
	pi := PostIn{
		PostBody: PostBody{
			Name: name,
		},
		Collection: collectionID,
	}
	return pi
}

// NewPost constructs a Post from a PostIn
func NewPost(id int32, ci PostIn) Post {
	now := time.Now()
	c := Post{
		ID: MakePostID(id),
		PostIn: PostIn{
			PostBody:   ci.PostBody,
			Collection: ci.Collection,
		},
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}

// MakePostID builds a Post ID string from an integer ID
func MakePostID(id int32) string {
	return pathPrefix + fmt.Sprintf(PostIDFormat, id)
}
