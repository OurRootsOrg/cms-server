package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"gocloud.dev/blob"
	"gocloud.dev/pubsub"

	"github.com/ourrootsorg/cms-server/model"
)

const ImagesPrefix = "/images/%d/"

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
		return nil, NewError(err)
	}
	return &PostResult{Posts: posts}, nil
}

// GetPost holds the business logic around getting a Post
func (api API) GetPost(ctx context.Context, id uint32) (*model.Post, error) {
	post, err := api.postPersister.SelectOnePost(ctx, id)
	if err != nil {
		return nil, NewError(err)
	}
	return post, nil
}

// AddPost holds the business logic around adding a Post
func (api API) AddPost(ctx context.Context, in model.PostIn) (*model.Post, error) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid post %v", err)
		return nil, NewError(err)
	}
	log.Printf("[DEBUG] Starting post %#v", in)
	log.Printf("[DEBUG] Open recordswriter topic")
	// prepare to send a RecordsWriter message
	recordsWriterTopic, err := api.OpenTopic(ctx, "recordswriter")
	if err != nil {
		log.Printf("[ERROR] Can't open recordswriter topic %v", err)
		return nil, NewError(err)
	}
	defer recordsWriterTopic.Shutdown(ctx)
	log.Printf("[DEBUG] Open imageswriter topic")
	// prepare to send a ImagesWriter message
	imagesWriterTopic, err := api.OpenTopic(ctx, "imageswriter")
	if err != nil {
		log.Printf("[ERROR] Can't open imageswriter topic %v", err)
		return nil, NewError(err)
	}
	defer imagesWriterTopic.Shutdown(ctx)
	// set records status
	in.RecordsStatus = model.PostDraft
	if in.RecordsKey != "" {
		in.RecordsStatus = model.PostLoading
	}
	// set images status
	in.ImagesStatus = model.PostDraft
	if len(in.ImagesKeys) > 0 {
		in.ImagesStatus = model.PostLoading
	}
	// insert
	post, e := api.postPersister.InsertPost(ctx, in)
	if e != nil {
		return nil, NewError(e)
	}

	if in.RecordsKey != "" {
		// send a message to write the records
		msg := model.RecordsWriterMsg{
			PostID: post.ID,
		}
		body, err := json.Marshal(msg)
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
		err = recordsWriterTopic.Send(ctx, &pubsub.Message{Body: body})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the insert
			_ = api.postPersister.DeletePost(ctx, post.ID)
			return nil, NewError(err)
		}
	}

	if len(in.ImagesKeys) > 0 {
		// send a message to write the images
		msg := model.ImagesWriterMsg{
			PostID:  post.ID,
			NewZips: in.ImagesKeys,
		}
		body, err := json.Marshal(msg)
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
		err = imagesWriterTopic.Send(ctx, &pubsub.Message{Body: body})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the insert
			_ = api.postPersister.DeletePost(ctx, post.ID)
			return nil, NewError(err)
		}
		log.Printf("[DEBUG] Sent imageswriter message '%s'", string(body))
	}

	return post, nil
}

// UpdatePost holds the business logic around updating a Post
func (api API) UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewError(err)
	}

	// read current records status
	currPost, errs := api.GetPost(ctx, id)
	if errs != nil {
		return nil, errs
	}

	var recordsWriterTopic, imagesWriterTopic *pubsub.Topic
	var msg []byte

	if currPost.RecordsKey != in.RecordsKey {
		// handle records key change
		if currPost.RecordsStatus != in.RecordsStatus || in.RecordsStatus != model.PostDraft {
			return nil, NewError(fmt.Errorf("cannot change recordsKey unless recordsStatus is Draft; status is %s", currPost.RecordsStatus))
		}
		// prepare to send a message
		recordsWriterTopic, err = api.OpenTopic(ctx, "recordswriter")
		if err != nil {
			log.Printf("[ERROR] Can't open recordswriter topic %v", err)
			return nil, NewError(err)
		}
		defer recordsWriterTopic.Shutdown(ctx)

		in.RecordsStatus = model.PostLoading
		msg, err = json.Marshal(model.RecordsWriterMsg{
			PostID: id,
		})
		if err != nil {
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
	} else if currPost.RecordsStatus != in.RecordsStatus {
		// handle records status change
		switch {
		case (currPost.RecordsStatus == model.PostDraft && in.RecordsStatus == model.PostPublished) ||
			(currPost.RecordsStatus == model.PostPublished && in.RecordsStatus == model.PostDraft):
			// prepare to send a message
			recordsWriterTopic, err = api.OpenTopic(ctx, "publisher")
			if err != nil {
				log.Printf("[ERROR] Can't open publisher topic %v", err)
				return nil, NewError(err)
			}
			defer recordsWriterTopic.Shutdown(ctx)

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
				return nil, NewError(err)
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
			return nil, NewError(errors.New(msg))
		}
	}

	if !currPost.ImagesKeys.Equals(&in.ImagesKeys) {
		// handle images key change
		if currPost.ImagesStatus != in.ImagesStatus || in.ImagesStatus != model.PostDraft {
			return nil, NewError(fmt.Errorf("cannot change imagesKey unless imagesStatus is Draft; status is %s", currPost.ImagesStatus))
		}
		// prepare to send a message
		imagesWriterTopic, err = api.OpenTopic(ctx, "imageswriter")
		if err != nil {
			log.Printf("[ERROR] Can't open imageswriter topic %v", err)
			return nil, NewError(err)
		}
		defer imagesWriterTopic.Shutdown(ctx)

		in.ImagesStatus = model.PostLoading
		iwm := model.ImagesWriterMsg{
			PostID: id,
		}
		for _, ik := range in.ImagesKeys {
			if !currPost.ImagesKeys.Contains(ik) {
				iwm.NewZips = append(iwm.NewZips, ik)
			}
		}
		msg, err = json.Marshal(iwm)
		if err != nil {
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
	} else if currPost.ImagesStatus != in.ImagesStatus {
		// handle images status change
		switch {
		case (currPost.ImagesStatus == model.PostDraft && in.ImagesStatus == model.PostPublished) ||
			(currPost.ImagesStatus == model.PostPublished && in.ImagesStatus == model.PostDraft):
			// prepare to send a message
			imagesWriterTopic, err = api.OpenTopic(ctx, "publisher")
			if err != nil {
				log.Printf("[ERROR] Can't open publisher topic %v", err)
				return nil, NewError(err)
			}
			defer imagesWriterTopic.Shutdown(ctx)

			var action string
			if in.ImagesStatus == model.PostPublished {
				in.ImagesStatus = model.PostPublishing
				action = model.PublisherActionIndex
			} else {
				in.ImagesStatus = model.PostUnpublishing
				action = model.PublisherActionUnindex
			}
			msg, err = json.Marshal(model.PublisherMsg{
				Action: action,
				PostID: id,
			})
			if err != nil {
				log.Printf("[ERROR] Can't marshal message %v", err)
				return nil, NewError(err)
			}
		case currPost.ImagesStatus == model.PostPublishing && in.ImagesStatus == model.PostPublishComplete:
			in.ImagesStatus = model.PostPublished
		case currPost.ImagesStatus == model.PostUnpublishing && in.ImagesStatus == model.PostUnpublishComplete:
			in.ImagesStatus = model.PostDraft
		case currPost.ImagesStatus == model.PostLoading && in.ImagesStatus == model.PostLoadComplete:
			in.ImagesStatus = model.PostDraft
		default:
			msg := fmt.Sprintf("[ERROR] cannot change images status from %s to %s", currPost.ImagesStatus, in.ImagesStatus)
			log.Println(msg)
			return nil, NewError(errors.New(msg))
		}
	}

	post, e := api.postPersister.UpdatePost(ctx, id, in)
	if e != nil {
		return nil, NewError(e)
	}

	if recordsWriterTopic != nil && msg != nil {
		// send the message
		err = recordsWriterTopic.Send(ctx, &pubsub.Message{Body: msg})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the update
			_, _ = api.postPersister.UpdatePost(ctx, id, *currPost)
			return nil, NewError(err)
		}
	}

	if currPost.RecordsKey != "" && currPost.RecordsKey != in.RecordsKey {
		api.deleteReferencedContent(ctx, currPost.RecordsKey)
	}

	if imagesWriterTopic != nil && msg != nil {
		// send the message
		err = imagesWriterTopic.Send(ctx, &pubsub.Message{Body: msg})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the update
			_, _ = api.postPersister.UpdatePost(ctx, id, *currPost)
			return nil, NewError(err)
		}
	}

	for _, ik := range currPost.ImagesKeys {
		if !in.ImagesKeys.Contains(ik) {
			// Delete the ZIP file
			api.deleteReferencedContent(ctx, ik)
		}
	}

	return post, nil
}

// DeletePost holds the business logic around deleting a Post
func (api API) DeletePost(ctx context.Context, id uint32) error {
	log.Printf("[DEBUG] deleting %d", id)
	post, err := api.GetPost(ctx, id)
	if err != nil {
		return NewError(err)
	}
	if post.RecordsStatus != model.PostDraft {
		return NewError(fmt.Errorf("post(%d).RecordsStatus must be Draft; is %s", post.ID, post.RecordsStatus))
	}
	if post.ImagesStatus != model.PostDraft && post.ImagesStatus != "" /* backwards compatibility */ {
		return NewError(fmt.Errorf("post(%d).ImagesStatus must be Draft; is %s", post.ID, post.ImagesStatus))
	}
	log.Printf("[DEBUG] deleting records for %d", id)
	// delete records for post first so we don't have referential integrity errors
	if err := api.DeleteRecordsForPost(ctx, id); err != nil {
		return err
	}
	log.Printf("[DEBUG] deleting post %d", id)
	if err := api.postPersister.DeletePost(ctx, id); err != nil {
		return NewError(err)
	}
	log.Printf("[DEBUG] deleting content for %d", id)
	if post.RecordsKey != "" {
		api.deleteReferencedContent(ctx, post.RecordsKey)
	}
	if len(post.ImagesKeys) > 0 {
		log.Printf("[DEBUG] deleting images for %d", id)
		if err := api.deleteImages(ctx, id); err != nil {
			return NewError(err)
		}
		// Delete ZIP files
		for _, ik := range post.ImagesKeys {
			api.deleteReferencedContent(ctx, ik)
		}
	}
	return nil
}

func (api API) deleteReferencedContent(ctx context.Context, key string) {
	// delete records data
	bucket, err := api.OpenBucket(ctx)
	if err != nil {
		log.Printf("[INFO] Error calling OpenBucket while deleting content %v: %v", key, err)
	}
	defer bucket.Close()
	if err := bucket.Delete(ctx, key); err != nil {
		log.Printf("[INFO] error deleting content %v: %v\n", key, err)
	}
}

// deleteImagesForPost holds the business logic around deleting the images for a Post
func (api API) deleteImages(ctx context.Context, postID uint32) error {
	bucket, err := api.OpenBucket(ctx)
	if err != nil {
		log.Printf("[ERROR] OpenBucket %v\n", err)
		return NewError(err)
	}
	defer bucket.Close()
	li := bucket.List(&blob.ListOptions{
		Prefix: fmt.Sprintf(ImagesPrefix, postID),
	})
	for {
		obj, err := li.Next(ctx)
		if err == io.EOF {
			break
		} else if err != nil {
			return NewError(err)
		}
		log.Printf("[DEBUG] Deleting key %s", obj.Key)
		err = bucket.Delete(ctx, obj.Key)
		if err != nil {
			log.Printf("[ERROR] Error deleting key %s", obj.Key)
			return NewError(err)
		}
	}
	return nil
}
