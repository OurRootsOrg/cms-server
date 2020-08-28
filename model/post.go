package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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

// UserAcceptedPostRecordsStatus returns true if its argument is a valid records status
func UserAcceptedPostRecordsStatus(status string) bool {
	for _, s := range []string{PostLoading, PostDraft, PostPublishing, PostPublished, PostUnpublishing} {
		if s == status {
			return true
		}
	}
	return false
}

// UserAcceptedPostImagesStatus returns true if its argument is a valid images status
func UserAcceptedPostImagesStatus(status string) bool {
	for _, s := range []string{PostLoading, PostDraft} {
		if s == status {
			return true
		}
	}
	return false
}

// ImagesWriterMsg represents a message to initiate processing of an image upload
type ImagesWriterMsg struct {
	PostID  uint32   `json:"postId"`
	NewZips []string `json:"newZips"`
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

// StringSet represents a set of unique strings
type StringSet []string

// NewStringSet constructs a StringSet
func NewStringSet() StringSet {
	return StringSet(make([]string, 0))
}

// Add adds a string to a StringSet
func (ss *StringSet) Add(s string) bool {
	if ss.Contains(s) {
		return false
	}
	*ss = append(*ss, s)
	return true
}

// Contains indicates whether a StringSet contains a string
func (ss *StringSet) Contains(s string) bool {
	for _, f := range *ss {
		if f == s {
			return true
		}
	}
	return false
}

// UnmarshalJSON unmarshals JSON to a StringSet
func (ss *StringSet) UnmarshalJSON(b []byte) error {
	if ss == nil {
		*ss = NewStringSet()
	}
	var strings []string
	err := json.Unmarshal(b, &strings)
	if err != nil {
		return err
	}
	for _, value := range strings {
		if !ss.Add(value) {
			return fmt.Errorf("Attempt to add duplicate string: %s", value)
		}
	}
	return nil
}

// Equals compares two StringSets and returns true if they contain the same strings
func (ss *StringSet) Equals(ss1 *StringSet) bool {
	if ss == ss1 {
		return true
	}
	for _, s := range *ss1 {
		if !ss.Contains(s) {
			return false
		}
	}
	for _, s := range *ss {
		if !ss1.Contains(s) {
			return false
		}
	}
	return true
}

// PostBody is the JSON body of a Post
// TODO Consider having a PostStatus which can be draft, publishing, published, or unpublishing
// TODO and instead of RecordsStatus and ImagesStatus, have RecordsLoading (bool) and ImagesLoading (bool)
// TODO Also consider having a RecordsLoadingError and ImagesLoadingError with the error result from the last load
type PostBody struct {
	Name          string                 `json:"name" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata"`
	RecordsKey    string                 `json:"recordsKey"`
	RecordsStatus string                 `json:"recordsStatus"`
	ImagesKeys    StringSet              `json:"imagesKeys"`
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
