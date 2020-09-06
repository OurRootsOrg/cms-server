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
// Draft (initial state) -> ToPublish
//   user can update Draft to ToPublish if RecordsStatus and ImagesStatus are both Default
//   this causes server to send an Index message to Publisher
// ToPublish -> Publishing
//   Publisher updates status to Publishing when starting to index
// Publishing -> Published or Error
//   Publisher updates status to Published if successful, or Error otherwise and sets PostError to the error message
// Published -> ToUnpublish
//   user can update Published to ToUnpublish
//   this causes server to send an Unindex message to Publisher
// ToUnpublish -> Unpublishing
//   Publisher updates status to Unpublishing when starting to unindex
// Unpublishing -> Draft or Error
//   Publisher updates status to to Draft if successful, or Error otherwise and sets PostError to the error message
// Error -> ToPublish
//   user can update Error to ToPublish if RecordsStatus and ImagesStatus are both Default
// Post can be deleted only in Draft or Error states and only when Records/Images statuses are in Default or Error states

type PostStatus string

const (
	PostStatusDraft        PostStatus = "Draft"
	PostStatusToPublish               = "Publication Requested"
	PostStatusPublishing              = "Publishing"
	PostStatusPublished               = "Published"
	PostStatusToUnpublish             = "Unpublication Requested"
	PostStatusUnpublishing            = "Unpublishing"
	PostStatusError                   = "Error"
)

// Records and Images statuses
// Default (initial state) -> ToLoad
//   when user adds a file to load, server sets status to ToLoad and sends a message to Images/Records Writer
// ToLoad -> Loading
//   Images/Records Writer updates status to Loading when starting to load
// Loading -> Default or Error
//   Images/Records Writer updates status to Default if successful or Error otherwise and sets Images/Records Error to the error message
// Error -> ToLoad
//   when user updates the file to load, server sets status to ToLoad and sends a message to Images/Records Writer
// Post can be deleted only when Images/Records status is Default or Error

type RecordsStatus string

const (
	RecordsStatusDefault RecordsStatus = ""
	RecordsStatusToLoad                = "Load Requested"
	RecordsStatusLoading               = "Loading"
	RecordsStatusError                 = "Error"
)

type ImagesStatus string

const (
	ImagesStatusDefault ImagesStatus = ""
	ImagesStatusToLoad               = "Load Requested"
	ImagesStatusLoading              = "Loading"
	ImagesStatusError                = "Error"
)

// Publisher actions
type PublisherAction string

const (
	PublisherActionIndex   PublisherAction = "index"
	PublisherActionUnindex                 = "unindex"
)

// UserAcceptedPostStatus returns true if status can be submitted by a user
func UserAcceptedPostStatus(status PostStatus) bool {
	for _, s := range []PostStatus{PostStatusDraft, PostStatusToPublish, PostStatusPublished, PostStatusToUnpublish, PostStatusError} {
		if s == status {
			return true
		}
	}
	return false
}

// UserAcceptedPostRecordsStatus returns true if status can be submitted by a user
func UserAcceptedPostRecordsStatus(status RecordsStatus) bool {
	for _, s := range []RecordsStatus{RecordsStatusDefault, RecordsStatusError} {
		if s == status {
			return true
		}
	}
	return false
}

// UserAcceptedPostImagesStatus returns true if status can be submitted by a user
func UserAcceptedPostImagesStatus(status ImagesStatus) bool {
	for _, s := range []ImagesStatus{ImagesStatusDefault, ImagesStatusError} {
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

// PublisherMsg represents a message to initiate publishing of a post
type PublisherMsg struct {
	Action PublisherAction `json:"action"`
	PostID uint32          `json:"postId"`
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
type PostBody struct {
	Name          string                 `json:"name" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata"`
	PostStatus    PostStatus             `json:"postStatus"`
	PostError     string                 `json:"postError"`
	RecordsKey    string                 `json:"recordsKey"`
	RecordsStatus RecordsStatus          `json:"recordsStatus"`
	RecordsError  string                 `json:"recordsError"`
	ImagesKeys    StringSet              `json:"imagesKeys"`
	ImagesStatus  ImagesStatus           `json:"imagesStatus"`
	ImagesError   string                 `json:"imagesError"`
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
