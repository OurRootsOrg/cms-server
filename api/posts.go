package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/disintegration/imaging"
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
	// default ImagesStatus, but see the comment about ImagesLoading in model/post.go
	for i := range posts {
		if posts[i].ImagesStatus == "" {
			posts[i].ImagesStatus = model.PostDraft
		}
	}
	return &PostResult{Posts: posts}, nil
}

// GetPost holds the business logic around getting a Post
func (api API) GetPost(ctx context.Context, id uint32) (*model.Post, error) {
	post, err := api.postPersister.SelectOnePost(ctx, id)
	if err != nil {
		return nil, NewError(err)
	}
	// default ImagesStatus, but see the comment about ImagesLoading in model/post.go
	if post.ImagesStatus == "" {
		post.ImagesStatus = model.PostDraft
	}
	return post, nil
}

const thumbQuality = 75

// GetPostImage returns a signed S3 URL to return an image file
func (api *API) GetPostImage(ctx context.Context, id uint32, filePath string, expireSeconds, height, width int) (*ImageMetadata, error) {
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

	// return full image?
	if height == 0 && width == 0 {
		// kind of a shame to read and decode the entire image just to get the dimensions
		// we read the image once and could store the dimensions in a separate _dimensions.json file for future use,
		// but this seems like a case of pre-mature optimization
		// openseadragon, used by the client, needs the image dimensions in order to function properly
		reader, err := bucket.NewReader(ctx, key, nil)
		if err != nil {
			log.Printf("[ERROR] GetPostImage read image %#v\n", err)
			return nil, NewError(fmt.Errorf("GetPostImage read image %v", err))
		}
		defer reader.Close()
		img, err := imaging.Decode(reader, imaging.AutoOrientation(true))
		if err != nil {
			log.Printf("[ERROR] GetPostImage decode %#v\n", err)
			return nil, NewError(fmt.Errorf("GetPostImage decode image %v", err))
		}

		signedURL, err := signingBucket.SignedURL(ctx, key, &blob.SignedURLOptions{
			Expiry: time.Duration(expireSeconds) * time.Second,
			Method: "GET",
		})
		if err != nil {
			return nil, NewError(err)
		}
		log.Printf("SignedURL %s\n", signedURL)
		return &ImageMetadata{
			URL:    signedURL,
			Height: img.Bounds().Dy(),
			Width:  img.Bounds().Dx(),
		}, nil
	}

	// generate and return a thumbnail
	thumbKey := fmt.Sprintf("%s__thumb_%dx%d", key, height, width)

	// does the thumbnail already exist?
	exists, err := bucket.Exists(ctx, thumbKey)
	if err != nil {
		log.Printf("[ERROR] GetPostImage check thumb exists %#v\n", err)
		return nil, NewError(fmt.Errorf("GetPostImage exists error %v", err))
	}

	// generate the thumbnail and save it
	if !exists {
		reader, err := bucket.NewReader(ctx, key, nil)
		if err != nil {
			log.Printf("[ERROR] GetPostImage read image %#v\n", err)
			return nil, NewError(fmt.Errorf("GetPostImage read image %v", err))
		}
		defer reader.Close()
		img, err := imaging.Decode(reader, imaging.AutoOrientation(true))
		if err != nil {
			log.Printf("[ERROR] GetPostImage decode %#v\n", err)
			return nil, NewError(fmt.Errorf("GetPostImage decode image %v", err))
		}
		// scale image
		thumb := imaging.Resize(img, width, height, imaging.Box)

		// write image
		writer, err := bucket.NewWriter(ctx, thumbKey, &blob.WriterOptions{
			ContentType: "image/jpeg",
		})
		if err != nil {
			log.Printf("[ERROR] GetPostImage start write image %#v\n", err)
			return nil, NewError(fmt.Errorf("GetPostImage start write image %v", err))
		}
		err = imaging.Encode(writer, thumb, imaging.JPEG, imaging.JPEGQuality(thumbQuality))
		closeErr := writer.Close()
		if err != nil || closeErr != nil {
			log.Printf("[ERROR] GetPostImage write image %#v close %#v\n", err, closeErr)
			return nil, NewError(fmt.Errorf("GetPostImage write image %v close %v", err, closeErr))
		}
	}

	// return the thumbnail
	signedURL, err := signingBucket.SignedURL(ctx, thumbKey, &blob.SignedURLOptions{
		Expiry: time.Duration(expireSeconds) * time.Second,
		Method: "GET",
	})
	if err != nil {
		return nil, NewError(err)
	}
	log.Printf("SignedURL %s\n", signedURL)
	return &ImageMetadata{
		URL:    signedURL,
		Height: height,
		Width:  width,
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

	var recordsWriterTopic, imagesWriterTopic, publisherTopic *pubsub.Topic
	var recordsMsg, imagesMsg, publisherMsg []byte

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
		recordsMsg, err = json.Marshal(model.RecordsWriterMsg{
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
			publisherTopic, err = api.OpenTopic(ctx, "publisher")
			if err != nil {
				log.Printf("[ERROR] Can't open publisher topic %v", err)
				return nil, NewError(err)
			}
			defer publisherTopic.Shutdown(ctx)

			var action string
			if in.RecordsStatus == model.PostPublished {
				in.RecordsStatus = model.PostPublishing
				action = model.PublisherActionIndex
			} else {
				in.RecordsStatus = model.PostUnpublishing
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
		imagesMsg, err = json.Marshal(iwm)
		if err != nil {
			log.Printf("[ERROR] Can't marshal message %v", err)
			return nil, NewError(err)
		}
	} else if currPost.ImagesStatus != in.ImagesStatus {
		// handle images status change
		switch {
		// images don't need to be published; only loaded
		case currPost.ImagesStatus == model.PostLoading && in.ImagesStatus == model.PostLoadComplete:
			in.ImagesStatus = model.PostDraft
		default:
			msg := fmt.Sprintf("[ERROR] cannot change images status from %s to %s", currPost.ImagesStatus, in.ImagesStatus)
			log.Println(msg)
			return nil, NewError(errors.New(msg))
		}
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

	if currPost.RecordsKey != "" && currPost.RecordsKey != in.RecordsKey {
		api.deleteReferencedContent(ctx, currPost.RecordsKey)
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

	for _, ik := range currPost.ImagesKeys {
		if !in.ImagesKeys.Contains(ik) {
			// Delete the ZIP file
			api.deleteReferencedContent(ctx, ik)
		}
	}

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
	// allow deleting posts where the records or images are stuck in loading status
	if post.RecordsStatus == model.PostPublished || post.RecordsStatus == model.PostPublishing || post.RecordsStatus == model.PostUnpublishing {
		return NewError(fmt.Errorf("post(%d).RecordsStatus must be Draft; is %s", post.ID, post.RecordsStatus))
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
			log.Printf("[ERROR] deleting images for %d error=%v", id, err)
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
	bucket, err := api.OpenBucket(ctx, false)
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
	bucket, err := api.OpenBucket(ctx, false)
	if err != nil {
		log.Printf("[ERROR] OpenBucket %#v\n", err)
		return NewError(err)
	}
	defer bucket.Close()
	prefix := fmt.Sprintf(ImagesPrefix, postID)
	li := bucket.List(&blob.ListOptions{
		Prefix: prefix,
	})
	for {
		obj, err := li.Next(ctx)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("[ERROR] Error getting next object with prefix %s: %#v", prefix, err)
			return NewError(err)
		}
		log.Printf("[DEBUG] Deleting key %s", obj.Key)
		err = bucket.Delete(ctx, obj.Key)
		if err != nil {
			log.Printf("[ERROR] Error deleting key %s: %#v", obj.Key, err)
			return NewError(err)
		}
	}
	return nil
}
