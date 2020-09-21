package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ourrootsorg/cms-server/model"
	"gocloud.dev/blob"
	"gocloud.dev/pubsub"
)

const ImagesPrefix = "/images/%d/"

// PostResult is a paged Post result
type PostResult struct {
	Posts    []model.Post `json:"posts"`
	NextPage string       `json:"next_page"`
}

type ImageMetadata struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
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

// GetPostImage returns a signed S3 URL to return an image file
func (api *API) GetPostImage(ctx context.Context, id uint32, filePath string, thumbnail bool, expireSeconds int) (*ImageMetadata, error) {
	// need an external bucket for signing so signed URLs can be used externally
	signingBucket, err := api.OpenBucket(ctx, true)
	if err != nil {
		return nil, NewError(err)
	}
	defer signingBucket.Close()

	bucket, err := api.OpenBucket(ctx, false)
	if err != nil {
		return nil, NewError(err)
	}
	defer bucket.Close()

	key := fmt.Sprintf(ImagesPrefix, id) + filePath
	if thumbnail {
		key += model.ImageThumbnailSuffix
	}
	dimensionsKey := key + model.ImageDimensionsSuffix

	// read image dimensions
	reader, err := bucket.NewReader(ctx, dimensionsKey, nil)
	if err != nil {
		log.Printf("[ERROR] GetPostImage read image %#v\n", err)
		return nil, NewError(fmt.Errorf("GetPostImage read image %v", err))
	}
	defer reader.Close()
	dec := json.NewDecoder(reader)
	var dim model.ImageDimensions
	if err := dec.Decode(&dim); err != nil {
		log.Printf("[ERROR] GetPostImage read dimensions %#v\n", err)
		return nil, NewError(fmt.Errorf("GetPostImage read dimensions %v", err))
	}

	// generate signed URL and return
	signedURL, err := signingBucket.SignedURL(ctx, key, &blob.SignedURLOptions{
		Expiry: time.Duration(expireSeconds) * time.Second,
		Method: "GET",
	})
	if err != nil {
		return nil, NewError(err)
	}
	return &ImageMetadata{
		URL:    signedURL,
		Height: dim.Height,
		Width:  dim.Width,
	}, nil
}

// AddPost holds the business logic around adding a Post
func (api API) AddPost(ctx context.Context, in model.PostIn) (*model.Post, error) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid post %v", err)
		return nil, NewError(err)
	}
	log.Printf("[DEBUG] Starting post %#v", in)
	in.PostStatus = model.PostStatusDraft
	in.RecordsStatus = model.RecordsStatusDefault
	in.ImagesStatus = model.ImagesStatusDefault

	var recordsWriterTopic, imagesWriterTopic *pubsub.Topic
	if in.RecordsKey != "" {
		log.Printf("[DEBUG] Open recordswriter topic")
		// prepare to send a RecordsWriter message
		recordsWriterTopic, err = api.OpenTopic(ctx, "recordswriter")
		if err != nil {
			log.Printf("[ERROR] Can't open recordswriter topic %v", err)
			return nil, NewError(err)
		}
		defer recordsWriterTopic.Shutdown(ctx)
		in.RecordsStatus = model.RecordsStatusToLoad
	}
	if len(in.ImagesKeys) > 0 {
		log.Printf("[DEBUG] Open imageswriter topic")
		// prepare to send a ImagesWriter message
		imagesWriterTopic, err = api.OpenTopic(ctx, "imageswriter")
		if err != nil {
			log.Printf("[ERROR] Can't open imageswriter topic %v", err)
			return nil, NewError(err)
		}
		defer imagesWriterTopic.Shutdown(ctx)
		in.ImagesStatus = model.ImagesStatusToLoad
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
			Action:  model.ImagesWriterActionUnzip,
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

	var recordsWriterTopic, imagesWriterTopic, publisherTopic *pubsub.Topic
	var recordsMsg, imagesMsg, publisherMsg []byte

	// validate records status change
	switch {
	case currPost.RecordsStatus == in.RecordsStatus:
		// continue
	case (currPost.RecordsStatus == model.RecordsStatusToLoad || currPost.RecordsStatus == model.RecordsStatusError) && in.RecordsStatus == model.RecordsStatusLoading:
		// continue
	case currPost.RecordsStatus == model.RecordsStatusLoading && in.RecordsStatus == model.RecordsStatusLoadComplete:
		in.RecordsStatus = model.RecordsStatusDefault
	case currPost.RecordsStatus == model.RecordsStatusLoading && in.RecordsStatus == model.RecordsStatusLoadError:
		in.RecordsStatus = model.RecordsStatusError
	default:
		err := fmt.Errorf("post %d cannot be updated from records status %s to status %s", currPost.ID, currPost.RecordsStatus, in.RecordsStatus)
		log.Printf("[DEBUG] %s", err.Error())
		return nil, NewHTTPError(err, http.StatusBadRequest)
	}

	// validate images status change
	switch {
	case currPost.ImagesStatus == in.ImagesStatus:
		// continue
	case (currPost.ImagesStatus == model.ImagesStatusToLoad || currPost.ImagesStatus == model.ImagesStatusError) && in.ImagesStatus == model.ImagesStatusLoading:
		// continue
	case currPost.ImagesStatus == model.ImagesStatusLoading && in.ImagesStatus == model.ImagesStatusLoadComplete:
		in.ImagesStatus = model.ImagesStatusDefault
	case currPost.ImagesStatus == model.ImagesStatusLoading && in.ImagesStatus == model.ImagesStatusLoadError:
		in.ImagesStatus = model.ImagesStatusError
	default:
		err := fmt.Errorf("post %d cannot be updated from images status %s to status %s", currPost.ID, currPost.ImagesStatus, in.ImagesStatus)
		log.Printf("[DEBUG] %s", err.Error())
		return nil, NewHTTPError(err, http.StatusBadRequest)
	}

	// handle records key change
	if currPost.RecordsKey != in.RecordsKey {
		if currPost.PostStatus != model.PostStatusDraft && currPost.PostStatus != model.PostStatusError {
			err := fmt.Errorf("cannot upload records for post %d unless post status is Draft or Error; status is %s", currPost.ID, currPost.PostStatus)
			log.Printf("[DEBUG] %s", err.Error())
			return nil, NewHTTPError(err, http.StatusBadRequest)
		}
		if currPost.RecordsStatus != model.RecordsStatusDefault && currPost.RecordsStatus != model.RecordsStatusError {
			err := fmt.Errorf("cannot upload records for post %d unless records status is empty or Error; status is %s", currPost.ID, currPost.RecordsStatus)
			log.Printf("[DEBUG] %s", err.Error())
			return nil, NewHTTPError(err, http.StatusBadRequest)
		}
		// prepare to send a message
		recordsWriterTopic, err = api.OpenTopic(ctx, "recordswriter")
		if err != nil {
			log.Printf("[ERROR] Can't open recordswriter topic %v", err)
			return nil, NewError(err)
		}
		defer recordsWriterTopic.Shutdown(ctx)

		in.RecordsStatus = model.RecordsStatusToLoad
		in.RecordsError = ""
		recordsMsg, err = json.Marshal(model.RecordsWriterMsg{
			PostID: id,
		})
		if err != nil {
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
	}

	// handle images key change
	if !currPost.ImagesKeys.Equals(&in.ImagesKeys) {
		if currPost.PostStatus != model.PostStatusDraft && currPost.PostStatus != model.PostStatusError {
			err := fmt.Errorf("cannot upload images for post %d unless post status is Draft or Error; status is %s", currPost.ID, currPost.PostStatus)
			log.Printf("[DEBUG] %s", err.Error())
			return nil, NewHTTPError(err, http.StatusBadRequest)

		}
		if currPost.ImagesStatus != model.ImagesStatusDefault && currPost.ImagesStatus != model.ImagesStatusError {
			err := fmt.Errorf("cannot upload images for post %d unless images status is empty or Error; status is %s", currPost.ID, currPost.ImagesStatus)
			log.Printf("[DEBUG] %s", err.Error())
			return nil, NewHTTPError(err, http.StatusBadRequest)
		}
		// prepare to send a message
		imagesWriterTopic, err = api.OpenTopic(ctx, "imageswriter")
		if err != nil {
			log.Printf("[ERROR] Can't open imageswriter topic %v", err)
			return nil, NewError(err)
		}
		defer imagesWriterTopic.Shutdown(ctx)

		in.ImagesStatus = model.ImagesStatusToLoad
		in.ImagesError = ""
		iwm := model.ImagesWriterMsg{
			Action: model.ImagesWriterActionUnzip,
			PostID: id,
		}
		for _, ik := range in.ImagesKeys {
			if !currPost.ImagesKeys.Contains(ik) {
				iwm.NewZips = append(iwm.NewZips, ik)
			}
		}
		imagesMsg, err = json.Marshal(iwm)
		if err != nil {
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
	}

	// validate post status change
	switch {
	case currPost.PostStatus == in.PostStatus:
		// continue
	case ((currPost.PostStatus == model.PostStatusDraft || currPost.PostStatus == model.PostStatusError) && in.PostStatus == model.PostStatusToPublish) ||
		(currPost.PostStatus == model.PostStatusPublished && in.PostStatus == model.PostStatusToUnpublish):
		if currPost.RecordsStatus != model.RecordsStatusDefault || currPost.ImagesStatus != model.ImagesStatusDefault {
			err := fmt.Errorf("post %d status can be Publication or Unpublication Requested only when records and images "+
				"statuses are empty; records status is %s and images status is %s", currPost.ID, currPost.RecordsStatus, currPost.ImagesStatus)
			log.Printf("[DEBUG] %s", err.Error())
			return nil, NewHTTPError(err, http.StatusBadRequest)
		}
		if in.RecordsKey == "" {
			err := fmt.Errorf("post %d status can be Publication Requested only when post has records; records key is empty", currPost.ID)
			log.Printf("[DEBUG] %s", err.Error())
			return nil, NewHTTPError(err, http.StatusBadRequest)
		}
		// prepare to send a message
		publisherTopic, err = api.OpenTopic(ctx, "publisher")
		if err != nil {
			log.Printf("[ERROR] Can't open publisher topic %v", err)
			return nil, NewError(err)
		}
		defer publisherTopic.Shutdown(ctx)

		in.PostError = ""
		var action model.PublisherAction
		if in.PostStatus == model.PostStatusToPublish {
			action = model.PublisherActionIndex
		} else {
			action = model.PublisherActionUnindex
		}
		publisherMsg, err = json.Marshal(model.PublisherMsg{
			Action: action,
			PostID: id,
		})
		if err != nil {
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
	case (currPost.PostStatus == model.PostStatusToPublish || currPost.PostStatus == model.PostStatusError) && in.PostStatus == model.PostStatusPublishing:
		// continue
	case currPost.PostStatus == model.PostStatusPublishing && in.PostStatus == model.PostStatusPublishComplete:
		in.PostStatus = model.PostStatusPublished
	case currPost.PostStatus == model.PostStatusPublishing && in.PostStatus == model.PostStatusPublishError:
		in.PostStatus = model.PostStatusError
	case (currPost.PostStatus == model.PostStatusToUnpublish || currPost.PostStatus == model.PostStatusError) && in.PostStatus == model.PostStatusUnpublishing:
		// continue
	case currPost.PostStatus == model.PostStatusUnpublishing && in.PostStatus == model.PostStatusUnpublishComplete:
		in.PostStatus = model.PostStatusDraft
	case currPost.PostStatus == model.PostStatusUnpublishing && in.PostStatus == model.PostStatusUnpublishError:
		in.PostStatus = model.PostStatusError
	default:
		err := fmt.Errorf("post %d cannot be updated from status %s to status %s", currPost.ID, currPost.PostStatus, in.PostStatus)
		log.Printf("[DEBUG] %s", err.Error())
		return nil, NewHTTPError(err, http.StatusBadRequest)
	}

	// update post
	post, e := api.postPersister.UpdatePost(ctx, id, in)
	if e != nil {
		return nil, NewError(e)
	}

	// send message to records writer
	if recordsWriterTopic != nil && recordsMsg != nil {
		// send the message
		err = recordsWriterTopic.Send(ctx, &pubsub.Message{Body: recordsMsg})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the update
			_, _ = api.postPersister.UpdatePost(ctx, id, *currPost)
			return nil, NewError(err)
		}
	}

	// send message to images writer
	if imagesWriterTopic != nil && imagesMsg != nil {
		// send the message
		err = imagesWriterTopic.Send(ctx, &pubsub.Message{Body: imagesMsg})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the update
			_, _ = api.postPersister.UpdatePost(ctx, id, *currPost)
			return nil, NewError(err)
		}
	}

	// send message to publisher
	if publisherTopic != nil && publisherMsg != nil {
		// send the message
		err = publisherTopic.Send(ctx, &pubsub.Message{Body: publisherMsg})
		if err != nil { // this had best never happen
			log.Printf("[ERROR] Can't send message %v", err)
			// undo the update
			_, _ = api.postPersister.UpdatePost(ctx, id, *currPost)
			return nil, NewError(err)
		}
	}

	// remove old records if any
	if currPost.RecordsKey != "" && currPost.RecordsKey != in.RecordsKey {
		if err := api.deleteReferencedContent(ctx, currPost.RecordsKey); err != nil {
			// log the error but don't undo the update
			log.Printf("[ERROR] deleting records %s when updating post %d %v", currPost.RecordsKey, currPost.ID, err)
		}
	}

	// remove old image zips if any
	for _, ik := range currPost.ImagesKeys {
		if !in.ImagesKeys.Contains(ik) {
			// Delete the ZIP file
			if err := api.deleteReferencedContent(ctx, ik); err != nil {
				// log the error but don't undo the update
				log.Printf("[ERROR] deleting zip file %s when updating post %d %v", ik, currPost.ID, err)
			}
		}
	}

	return post, nil
}

// DeletePost holds the business logic around deleting a Post
func (api API) DeletePost(ctx context.Context, id uint32) error {
	log.Printf("[DEBUG] deleting %d", id)
	post, err := api.GetPost(ctx, id)
	if err != nil {
		log.Printf("[ERROR] reading post %d error=%v", id, err)
		return NewError(err)
	}
	// allow deleting posts only when post is draft or error, and when records and images are default or error
	if (post.PostStatus != model.PostStatusDraft && post.PostStatus != model.PostStatusError) ||
		(post.RecordsStatus != model.RecordsStatusDefault && post.RecordsStatus != model.RecordsStatusError) ||
		(post.ImagesStatus != model.ImagesStatusDefault && post.ImagesStatus != model.ImagesStatusError) {
		return NewError(fmt.Errorf("post %d status must be Draft or Error, and records and images statuses must be empty or Error; "+
			"post status is %s, records status is %s, and images status is %s", post.ID, post.PostStatus, post.RecordsStatus, post.ImagesStatus))
	}

	log.Printf("[DEBUG] deleting records for %d", id)
	// delete record households for post first so we don't have referential integrity errors
	if err := api.DeleteRecordHouseholdsForPost(ctx, id); err != nil {
		return err
	}
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
		if err := api.deleteReferencedContent(ctx, post.RecordsKey); err != nil {
			// log the error but don't undo the delete
			log.Printf("[ERROR] deleting records %s when deleting post %d %v", post.RecordsKey, post.ID, err)
		}
	}
	if len(post.ImagesKeys) > 0 {
		log.Printf("[DEBUG] deleting images for %d", id)
		if err := api.deleteImages(ctx, id); err != nil {
			// log the error but don't undo the delete
			log.Printf("[ERROR] deleting images when deleting post %d %v", post.ID, err)
		}
		// Delete ZIP files
		for _, ik := range post.ImagesKeys {
			if err := api.deleteReferencedContent(ctx, ik); err != nil {
				// log the error but don't undo the delete
				log.Printf("[ERROR] deleting zip file %s when deleting post %d %v", ik, post.ID, err)
			}
		}
	}

	return nil
}

func (api API) deleteReferencedContent(ctx context.Context, key string) error {
	// delete records data
	bucket, err := api.OpenBucket(ctx, false)
	if err != nil {
		return err
	}
	defer bucket.Close()
	if err := bucket.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

// deleteImagesForPost holds the business logic around deleting the images for a Post
func (api API) deleteImages(ctx context.Context, postID uint32) error {
	bucket, err := api.OpenBucket(ctx, false)
	if err != nil {
		return err
	}
	defer bucket.Close()
	prefix := fmt.Sprintf(ImagesPrefix, postID)
	li := bucket.List(&blob.ListOptions{
		Prefix: prefix,
	})
	var errs []string
	for {
		obj, e := li.Next(ctx)
		if e == io.EOF {
			break
		} else if e != nil {
			errs = append(errs, fmt.Sprintf("error getting next object with prefix %s: %v", prefix, e))
		} else {
			log.Printf("[DEBUG] Deleting key %s", obj.Key)
			e = bucket.Delete(ctx, obj.Key)
			if e != nil {
				errs = append(errs, fmt.Sprintf("error deleting key %s: %v", obj.Key, e))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("error(s) deleting images: %s", strings.Join(errs, "; "))
	}
	return nil
}
