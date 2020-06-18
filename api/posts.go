package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gocloud.dev/pubsub"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
)

const PostLoading = "Loading"
const PostDraft = "Draft"
const PostPublishing = "Publishing"
const PostPublished = "Published"
const PostPublishComplete = "PublishComplete"

type RecordsWriterMsg struct {
	PostID uint32 `json:"postId"`
}

type PublisherMsg struct {
	PostID uint32 `json:"postId"`
}

// PostResult is a paged Post result
type PostResult struct {
	Posts    []model.Post `json:"posts"`
	NextPage string       `json:"next_page"`
}

// GetPosts holds the business logic around getting many Posts
func (api API) GetPosts(ctx context.Context /* filter/search criteria */) (*PostResult, *model.Errors) {
	// TODO: handle search criteria and paged results
	posts, err := api.postPersister.SelectPosts(ctx)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &PostResult{Posts: posts}, nil
}

// GetPost holds the business logic around getting a Post
func (api API) GetPost(ctx context.Context, id uint32) (*model.Post, *model.Errors) {
	post, err := api.postPersister.SelectOnePost(ctx, id)
	if err == persist.ErrNoRows {
		return nil, model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, strconv.Itoa(int(id))))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &post, nil
}

// AddPost holds the business logic around adding a Post
func (api API) AddPost(ctx context.Context, in model.PostIn) (*model.Post, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid post %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	// prepare to send a message
	topic, err := api.OpenTopic(ctx, "recordswriter")
	if err != nil {
		log.Printf("[ERROR] Can't open recordswriter topic %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	defer topic.Shutdown(ctx)
	// set records status
	in.RecordsStatus = PostLoading
	// insert
	post, err := api.postPersister.InsertPost(ctx, in)
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid collection reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, strconv.Itoa(int(in.Collection)), "collection"))
	} else if err != nil {
		log.Printf("[ERROR] Internal server error: %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	// send a message to write the records
	msg := RecordsWriterMsg{
		PostID: post.ID,
	}
	body, err := json.Marshal(msg)
	if err != nil { // this had best never happen
		log.Printf("[ERROR] Can't marshal message %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	err = topic.Send(ctx, &pubsub.Message{Body: body})
	if err != nil { // this had best never happen
		log.Printf("[ERROR] Can't send message %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &post, nil
}

// UpdatePost holds the business logic around updating a Post
func (api API) UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}

	// read current records status
	currPost, errs := api.GetPost(ctx, id)
	if errs != nil {
		return nil, errs
	}

	// validate new records status
	var topic *pubsub.Topic
	err = errors.New(fmt.Sprintf("cannot change records status from %s to %s", currPost.RecordsStatus, in.RecordsStatus))
	if currPost.RecordsStatus == in.RecordsStatus {
		err = nil
	} else if currPost.RecordsStatus == PostDraft && in.RecordsStatus == PostPublished {
		// prepare to send a message
		topic, err = api.OpenTopic(ctx, "publisher")
		if err != nil {
			log.Printf("[ERROR] Can't open publisher topic %v", err)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		}
		defer topic.Shutdown(ctx)
		in.RecordsStatus = PostPublishing
		err = nil
	} else if currPost.RecordsStatus == PostPublishing && in.RecordsStatus == PostPublishComplete {
		in.RecordsStatus = PostPublished
		err = nil
	}

	post, err := api.postPersister.UpdatePost(ctx, id, in)
	if er, ok := err.(model.Error); ok {
		if er.Code == model.ErrConcurrentUpdate {
			return nil, model.NewErrors(http.StatusConflict, er)
		} else if er.Code == model.ErrNotFound {
			// Not allowed to add a Post with PUT
			return nil, model.NewErrors(http.StatusNotFound, er)
		}
	}
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid collection reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, strconv.Itoa(int(in.Collection)), "collection"))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}

	if currPost.RecordsStatus == PostDraft && in.RecordsStatus == PostPublishing {
		// send a message to publish the post
		msg := PublisherMsg{
			PostID: post.ID,
		}
		body, err := json.Marshal(msg)
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		}
		err = topic.Send(ctx, &pubsub.Message{Body: body})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		}
	}

	return &post, nil
}

// DeletePost holds the business logic around deleting a Post
func (api API) DeletePost(ctx context.Context, id uint32) *model.Errors {
	// delete records for post first so we don't have referential integrity errors
	if err := api.DeleteRecordsForPost(ctx, id); err != nil {
		return model.NewErrors(http.StatusInternalServerError, err)
	}
	if err := api.postPersister.DeletePost(ctx, id); err != nil {
		return model.NewErrors(http.StatusInternalServerError, err)
	}

	return nil
}
