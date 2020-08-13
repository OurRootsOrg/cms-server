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

// Post statuses
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

// Publisher actions
const (
	PublisherActionIndex   = "index"
	PublisherActionUnindex = "unindex"
)

// UserAcceptedPostRecordsStatus returns true if its argument is a valid post status
func UserAcceptedPostRecordsStatus(status string) bool {
	for _, s := range []string{PostLoading, PostDraft, PostPublishing, PostPublished, PostUnpublishing} {
		if s == status {
			return true
		}
	}
	return false
}

// ImagesWriterMsg represents a message to initiate processing of an image upload
type ImagesWriterMsg struct {
	PostID uint32 `json:"postId"`
}

// RecordsWriterMsg represents a message to initiate processing of an uploaded recods CSV
type RecordsWriterMsg struct {
	PostID uint32 `json:"postId"`
}

// PublisherMsg represents a message to initiate pulishing of a post
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
	ImagesKey     string                 `json:"imagesKey"`
	ImagesStatus  string                 `json:"imagesStatus"`
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
	Collection uint32 `json:"collection" example:"999" validate:"required" dynamodbav:"altSort,string"`
}

// Post represents a set of related Records
type Post struct {
	ID   uint32 `json:"id,omitempty" example:"999" validate:"required" dynamodbav:"pk"`
	Type string `json:"-" dynamodbav:"sk"`
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
