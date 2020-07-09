package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// PostPersister defines methods needed to persist categories
type PostPersister interface {
	SelectPosts(ctx context.Context) ([]Post, error)
	SelectOnePost(ctx context.Context, id uint32) (*Post, error)
	InsertPost(ctx context.Context, in PostIn) (*Post, error)
	UpdatePost(ctx context.Context, id uint32, in Post) (*Post, error)
	DeletePost(ctx context.Context, id uint32) error
}

const (
	PostLoading           = "Loading"
	PostLoadComplete      = "LoadComplete"
	PostDraft             = "Draft"
	PostPublishing        = "Publishing"
	PostPublished         = "Published"
	PostPublishComplete   = "PublishComplete" // set only by publisher
	PostUnpublishing      = "Unpublishing"
	PostUnpublishComplete = "UnpublishComplete" // set only by publisher
)
const (
	PublisherActionIndex   = "index"
	PublisherActionUnindex = "unindex"
)

func UserAcceptedPostRecordsStatus(status string) bool {
	for _, s := range []string{PostLoading, PostDraft, PostPublishing, PostPublished, PostUnpublishing} {
		if s == status {
			return true
		}
	}
	return false
}

type RecordsWriterMsg struct {
	PostID uint32 `json:"postId"`
}

type PublisherMsg struct {
	Action string `json:"action"`
	PostID uint32 `json:"postId"`
}

// PostBody is the JSON body of a Post
type PostBody struct {
	Name          string                 `json:"name" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata"`
	RecordsKey    string                 `json:"recordsKey"`
	RecordsStatus string                 `json:"recordsStatus"`
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
	Collection uint32 `json:"collection" example:"999" validate:"required"`
}

// Post represents a set of related Records
type Post struct {
	ID uint32 `json:"id,omitempty" example:"999" validate:"required"`
	PostIn
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewPostIn constructs a PostIn
func NewPostIn(name string, collectionID uint32, recordsKey string) PostIn {
	pi := PostIn{
		PostBody: PostBody{
			Name:       name,
			RecordsKey: recordsKey,
		},
		Collection: collectionID,
	}
	return pi
}

// NewPost constructs a Post from a PostIn
func NewPost(id uint32, ci PostIn) Post {
	now := time.Now()
	c := Post{
		ID: id,
		PostIn: PostIn{
			PostBody:   ci.PostBody,
			Collection: ci.Collection,
		},
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return c
}
