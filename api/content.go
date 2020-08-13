package api

import (
	"context"
	"fmt"
	"mime"
	"time"

	"gocloud.dev/blob"
)

// ContentRequest contains a content type of the content to post
type ContentRequest struct {
	ContentType string `json:"contentType"`
}

// ContentResult contains a URL to post the content
type ContentResult struct {
	PutURL string `json:"putURL"`
	Key    string `json:"key"`
}

// PostContentRequest returns a URL for posting content
func (api API) PostContentRequest(ctx context.Context, contentRequest ContentRequest) (*ContentResult, error) {
	mimeType, _, err := mime.ParseMediaType(contentRequest.ContentType)
	if err != nil && err != mime.ErrInvalidMediaParameter {
		return nil, NewError(fmt.Errorf("Bad ContentType %s: %v", contentRequest.ContentType, err))
	}
	bucket, err := api.OpenBucket(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	defer bucket.Close()

	now := time.Now()
	key := fmt.Sprintf("%s/%s", now.Format("2006-01-02"), now.Format(time.RFC3339Nano))
	// Maybe we should do this for other types, but that wouldn't be backwards compatible
	if mimeType == "application/zip" || mimeType == "application/x-zip-compressed" {
		key += ".zip"
	}
	signedURL, err := bucket.SignedURL(ctx, key, &blob.SignedURLOptions{
		Expiry:      5 * time.Minute,
		Method:      "PUT",
		ContentType: contentRequest.ContentType,
	})
	if err != nil {
		return nil, NewError(err)
	}
	return &ContentResult{signedURL, key}, nil
}

func (api API) GetContent(ctx context.Context, key string) ([]byte, error) {
	bucket, err := api.OpenBucket(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	defer bucket.Close()

	content, err := bucket.ReadAll(ctx, key)

	if err != nil {
		return nil, NewError(err)
	}
	return content, nil
}
