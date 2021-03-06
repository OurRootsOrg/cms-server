package api

import (
	"context"
	"fmt"
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
	bucket, err := api.OpenBucket(ctx, true)
	if err != nil {
		return nil, NewError(err)
	}
	defer bucket.Close()

	now := time.Now()
	key := fmt.Sprintf("%s/%s", now.Format("2006-01-02"), now.Format(time.RFC3339Nano))
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
	bucket, err := api.OpenBucket(ctx, false)
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
