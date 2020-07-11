package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"gocloud.dev/pubsub"

	"github.com/ourrootsorg/cms-server/model"
)

// PostResult is a paged Post result
type PostResult struct {
	Posts    []model.Post `json:"posts"`
	NextPage string       `json:"next_page"`
}

// GetPosts holds the business logic around getting many Posts
func (api API) GetPosts(ctx context.Context /* filter/search criteria */) (*PostResult, error) {
	// TODO: handle search criteria and paged results
	posts, err := api.postPersister.SelectPosts(ctx)
	if err != nil {
		return nil, model.NewErrors(0, err)
	}
	return &PostResult{Posts: posts}, nil
}

// GetPost holds the business logic around getting a Post
func (api API) GetPost(ctx context.Context, id uint32) (*model.Post, error) {
	post, err := api.postPersister.SelectOnePost(ctx, id)
	if err != nil {
		return nil, model.NewErrors(0, err)
	}
	return post, nil
}

// AddPost holds the business logic around adding a Post
func (api API) AddPost(ctx context.Context, in model.PostIn) (*model.Post, error) {
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
	in.RecordsStatus = model.PostDraft
	if in.RecordsKey != "" {
		in.RecordsStatus = model.PostLoading
	}
	// insert
	post, e := api.postPersister.InsertPost(ctx, in)
	if e != nil {
		return nil, model.NewErrors(0, e)
	}

	if in.RecordsKey != "" {
		// send a message to write the records
		msg := model.RecordsWriterMsg{
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
			// undo the insert
			_ = api.postPersister.DeletePost(ctx, post.ID)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		}
	}

	return post, nil
}

// UpdatePost holds the business logic around updating a Post
func (api API) UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}

	// read current records status
	currPost, errs := api.GetPost(ctx, id)
	if errs != nil {
		return nil, errs
	}

	var topic *pubsub.Topic
	var msg []byte

	if currPost.RecordsKey != in.RecordsKey {
		// handle records key change
		if currPost.RecordsStatus != in.RecordsStatus || in.RecordsStatus != model.PostDraft {
			return nil, model.NewErrors(http.StatusBadRequest, errors.New(fmt.Sprintf("cannot change recordsKey unless recordsStatus is Draft; status is %s", currPost.RecordsStatus)))
		}
		// prepare to send a message
		topic, err = api.OpenTopic(ctx, "recordswriter")
		if err != nil {
			log.Printf("[ERROR] Can't open recordswriter topic %v", err)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		}
		defer topic.Shutdown(ctx)

		in.RecordsStatus = model.PostLoading
		msg, err = json.Marshal(model.RecordsWriterMsg{
			PostID: id,
		})
		if err != nil {
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		}
	} else if currPost.RecordsStatus != in.RecordsStatus {
		// handle records status change
		switch {
		case (currPost.RecordsStatus == model.PostDraft && in.RecordsStatus == model.PostPublished) ||
			(currPost.RecordsStatus == model.PostPublished && in.RecordsStatus == model.PostDraft):
			// prepare to send a message
			topic, err = api.OpenTopic(ctx, "publisher")
			if err != nil {
				log.Printf("[ERROR] Can't open publisher topic %v", err)
				return nil, model.NewErrors(http.StatusInternalServerError, err)
			}
			defer topic.Shutdown(ctx)

			var action string
			if in.RecordsStatus == model.PostPublished {
				in.RecordsStatus = model.PostPublishing
				action = model.PublisherActionIndex
			} else {
				in.RecordsStatus = model.PostUnpublishing
				action = model.PublisherActionUnindex
			}
			msg, err = json.Marshal(model.PublisherMsg{
				Action: action,
				PostID: id,
			})
			if err != nil {
				log.Printf("[ERROR] Can't marshal message %v", err)
				return nil, model.NewErrors(http.StatusInternalServerError, err)
			}
		case currPost.RecordsStatus == model.PostPublishing && in.RecordsStatus == model.PostPublishComplete:
			in.RecordsStatus = model.PostPublished
		case currPost.RecordsStatus == model.PostUnpublishing && in.RecordsStatus == model.PostUnpublishComplete:
			in.RecordsStatus = model.PostDraft
		case currPost.RecordsStatus == model.PostLoading && in.RecordsStatus == model.PostLoadComplete:
			in.RecordsStatus = model.PostDraft
		default:
			msg := fmt.Sprintf("[ERROR] cannot change records status from %s to %s", currPost.RecordsStatus, in.RecordsStatus)
			log.Println(msg)
			return nil, model.NewErrors(http.StatusBadRequest, errors.New(msg))
		}
	}

	post, e := api.postPersister.UpdatePost(ctx, id, in)
	if e != nil {
		return nil, model.NewErrors(0, e)
	}

	if topic != nil && msg != nil {
		// send the message
		err = topic.Send(ctx, &pubsub.Message{Body: msg})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the update
			_, _ = api.postPersister.UpdatePost(ctx, id, *currPost)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		}
	}

	if currPost.RecordsKey != "" && currPost.RecordsKey != in.RecordsKey {
		api.deleteRecordsData(ctx, currPost.RecordsKey)
	}
	return post, nil
}

// DeletePost holds the business logic around deleting a Post
func (api API) DeletePost(ctx context.Context, id uint32) error {
	post, err := api.GetPost(ctx, id)
	if err != nil {
		return model.NewErrors(http.StatusNotFound, err)
	}
	if post.RecordsStatus != model.PostDraft {
		return model.NewErrors(http.StatusBadRequest, fmt.Errorf("post must be in Draft status; is %s", post.RecordsStatus))
	}
	// delete records for post first so we don't have referential integrity errors
	if err := api.DeleteRecordsForPost(ctx, id); err != nil {
		return err
	}
	if err := api.postPersister.DeletePost(ctx, id); err != nil {
		return model.NewErrors(0, err)
	}
	if post.RecordsKey != "" {
		api.deleteRecordsData(ctx, post.RecordsKey)
	}
	return nil
}

func (api API) deleteRecordsData(ctx context.Context, key string) {
	// delete records data
	bucket, err := api.OpenBucket(ctx)
	if err != nil {
		log.Printf("[ERROR] OpenBucket %v\n", err)
	}
	defer bucket.Close()
	if err := bucket.Delete(ctx, key); err != nil {
		log.Printf("[ERROR] error deleting records file %v\n", err)
	}
}
