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
//   user can update Draft to ToPublish if RecordsStatus and ImagesStatus are both Default and RecordsKey is not empty
//   this causes server to send an Index message to Publisher
// ToPublish -> Publishing
//   Publisher updates status to Publishing when starting to index
// Publishing -> PublishComplete or PublishError
//   Publisher updates status to PublishComplete if successful, or PublishError otherwise and sets PostError to the error message
//   having intermediate statuses for PublishComplete and PublishError means that user cannot change an Publishing status,
//     since PublishComplete and PublishError fail UserAcceptedPostStatus()
// PublishComplete -> Published
//   when the server gets PublishComplete from Publisher, it sets status to Published
// PublishError -> Error
//   when the server gets PublishError from Publisher, it sets status to Error
// Published -> ToUnpublish
//   user can update Published to ToUnpublish
//   this causes server to send an Unindex message to Publisher
// ToUnpublish -> Unpublishing
//   Publisher updates status to Unpublishing when starting to unindex
// Unpublishing -> UnpublishComplete or UnpublishError
//   Publisher updates status to to UnpublishComplete if successful, or UnpublishError otherwise and sets PostError to the error message
//   having intermediate statuses for UnpublishComplete and UnpublishError means that user cannot change an Unpublishing status,
//     since UnpublishComplete and UnpublishError fail UserAcceptedPostStatus()
// UnpublishComplete -> Draft
//   when the server gets UnpublishComplete from Publisher, it sets status to Draft
// UnpublishError -> Error
//   when the server gets UnpublishError from Publisher, it sets status to Error
// Error -> ToPublish or Publishing or Unpublishing
//   user can update Error to ToPublish if RecordsStatus and ImagesStatus are both Default
//   Publisher can update error to Publishing or Unpublishing when retrying
// Publisher will process the message only when status is To(Un)Publish or Error (in case the previous invocation failed)
// Post can be deleted only in Draft or Error states and only when Records/Images statuses are in Default or Error states

type PostStatus string

const (
	PostStatusDraft             PostStatus = "Draft"
	PostStatusToPublish         PostStatus = "Publication Requested"
	PostStatusPublishing        PostStatus = "Publishing"
	PostStatusPublishComplete   PostStatus = "PublishComplete"
	PostStatusPublishError      PostStatus = "PublishError"
	PostStatusPublished         PostStatus = "Published"
	PostStatusToUnpublish       PostStatus = "Unpublication Requested"
	PostStatusUnpublishing      PostStatus = "Unpublishing"
	PostStatusUnpublishComplete PostStatus = "UnpublishComplete"
	PostStatusUnpublishError    PostStatus = "UnpublishError"
	PostStatusError             PostStatus = "Error"
)

// Records and Images statuses
// Default (initial state) -> ToLoad
//   when user adds a file to load, server sets status to ToLoad and sends a message to Images/Records Writer
//   post status must be Draft or Error
// ToLoad -> Loading
//   Images/Records Writer updates status to Loading when starting to load
// Loading -> LoadComplete or LoadError
//   Images/Records Writer updates status to LoadComplete if successful or Error otherwise and sets Images/Records Error to the error message
//   having intermediate statuses for LoadComplete and LoadError means that user cannot change a Loading status,
//     since LoadComplete and LoadError fail UserAcceptedRecordsStatus() and UserAcceptedImagesStatus()
// LoadComplete -> Default
//   when the server gets LoadComplete from Images/Records Writer, it sets status to Default
// LoadError -> Error
//   when the server gets LoadError from Images/Records Writer, it sets status to Error
// Error -> ToLoad or Loading
//   when user updates the file to load, server sets status to ToLoad and sends a message to Images/Records Writer
//   Images/Records Writer updates Error to Loading when retrying
// Images/Records Writer will process the message only when status is ToLoad or Error (in case the previous invocation failed)
// Post can be deleted only when Images/Records status is Default or Error

type RecordsStatus string

const (
	RecordsStatusDefault      RecordsStatus = ""
	RecordsStatusToLoad       RecordsStatus = "Load Requested"
	RecordsStatusLoading      RecordsStatus = "Loading"
	RecordsStatusLoadComplete RecordsStatus = "LoadComplete"
	RecordsStatusLoadError    RecordsStatus = "LoadError"
	RecordsStatusError        RecordsStatus = "Error"
)

type ImagesStatus string

const (
	ImagesStatusDefault      ImagesStatus = ""
	ImagesStatusToLoad       ImagesStatus = "Load Requested"
	ImagesStatusLoading      ImagesStatus = "Loading"
	ImagesStatusLoadComplete ImagesStatus = "LoadComplete"
	ImagesStatusLoadError    ImagesStatus = "LoadError"
	ImagesStatusError        ImagesStatus = "Error"
)

// Publisher actions
type PublisherAction string

const (
	PublisherActionIndex   PublisherAction = "index"
	PublisherActionUnindex PublisherAction = "unindex"
)

// ImageWriter actions
type ImagesWriterAction string

const (
	ImagesWriterActionUnzip             ImagesWriterAction = "unzip"
	ImagesWriterActionGenerateThumbnail ImagesWriterAction = "thumb"
)

const ImageDimensionsSuffix = "__dimensions.json"
const ImageThumbnailSuffix = "__thumbnail.jpg"
const ImageThumbnailQuality = 75
const ImageThumbnailWidth = 160
const ImageThumbnailHeight = 0

type ImageDimensions struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

// UserAcceptedPostStatus returns true if status can be submitted by a user
func UserAcceptedPostStatus(status PostStatus) bool {
	for _, s := range []PostStatus{PostStatusDraft, PostStatusToPublish, PostStatusPublished, PostStatusToUnpublish, PostStatusError} {
		if s == status {
			return true
		}
	}
	return false
}

// UserAcceptedRecordsStatus returns true if status can be submitted by a user
func UserAcceptedRecordsStatus(status RecordsStatus) bool {
	for _, s := range []RecordsStatus{RecordsStatusDefault, RecordsStatusError} {
		if s == status {
			return true
		}
	}
	return false
}

// UserAcceptedImagesStatus returns true if status can be submitted by a user
func UserAcceptedImagesStatus(status ImagesStatus) bool {
	for _, s := range []ImagesStatus{ImagesStatusDefault, ImagesStatusError} {
		if s == status {
			return true
		}
	}
	return false
}

// ImagesWriterMsg represents a message to initiate processing of an image upload
type ImagesWriterMsg struct {
	PostID    uint32             `json:"postId"`
	Action    ImagesWriterAction `json:"action"`
	ImagePath string             `json:"imagePath"`
	NewZips   []string           `json:"newZips"`
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
	ID   uint32 `json:"id,omitempty" example:"999" validate:"required" dynamodbav:"pk,string"`
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
			PostStatus: PostStatusDraft,
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
