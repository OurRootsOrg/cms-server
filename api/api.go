package api

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/ourrootsorg/cms-server/service"
	"gocloud.dev/blob"

	"gocloud.dev/pubsub"

	"github.com/go-playground/validator/v10"
	"github.com/ourrootsorg/cms-server/model"
)

// API is the container for the apilication
type API struct {
	categoryPersister   model.CategoryPersister
	collectionPersister model.CollectionPersister
	baseURL             url.URL
	validate            *validator.Validate
	blobStoreConfig     BlobStoreConfig
	pubSubConfig        PubSubConfig
}

// BlobStoreConfig contains configuration information for the blob store
type BlobStoreConfig struct {
	region     string
	endpoint   string
	accessKey  string
	secretKey  string
	bucket     string
	disableSSL bool
}

// PubSubConfig contains configuration information for the pub sub service
type PubSubConfig struct {
	region   string
	protocol string
	prefix   string
}

// NewAPI builds an API
func NewAPI() *API {
	api := &API{
		baseURL: url.URL{},
	}
	api.Validate(validator.New())
	return api
}

// BaseURL sets the base URL for the api
func (api *API) BaseURL(url url.URL) *API {
	api.baseURL = url
	return api
}

// Validate sets the validate object for the api
func (api *API) Validate(validate *validator.Validate) *API {
	// Return JSON tag name as Field() in errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	api.validate = validate
	return api
}

// CategoryPersister sets the CategoryPersister for the api
func (api *API) CategoryPersister(cp model.CategoryPersister) *API {
	api.categoryPersister = cp
	return api
}

// CollectionPersister sets the CollectionPersister for the api
func (api *API) CollectionPersister(cp model.CollectionPersister) *API {
	api.collectionPersister = cp
	return api
}

// BlobStoreConfig configures the blob store service
func (api *API) BlobStoreConfig(region, endpoint, accessKeyID, secretAccessKey, bucket string, disableSSL bool) *API {
	api.blobStoreConfig = BlobStoreConfig{region, endpoint, accessKeyID, secretAccessKey, bucket, disableSSL}
	return api
}

// PubSubConfig configures the pub-sub service
func (api *API) PubSubConfig(region, protocol, prefix string) *API {
	api.pubSubConfig = PubSubConfig{region, protocol, prefix}
	return api
}

// OpenBucket opens a blob storage bucket; Close() the bucket when you're done with it
func (api *API) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	return service.OpenBucket(ctx, api.blobStoreConfig.bucket, api.blobStoreConfig.region, api.blobStoreConfig.endpoint,
		api.blobStoreConfig.accessKey, api.blobStoreConfig.secretKey, api.blobStoreConfig.disableSSL)
}

// OpenTopic opens a topic for publishing
// Shutdown(ctx) the topic when you're done with it
func (api *API) OpenTopic(ctx context.Context, topic string) (*pubsub.Topic, error) {
	return pubsub.OpenTopic(ctx, api.getPubSubUrlStr(topic))
}

// Shutdown(ctx) the subscription when you're done with it, and ack() messages when you've processed them
func (api *API) OpenSubscription(ctx context.Context, queue string) (*pubsub.Subscription, error) {
	return pubsub.OpenSubscription(ctx, api.getPubSubUrlStr(queue))
}

func (api *API) getPubSubUrlStr(target string) string {
	var urlStr string
	switch api.pubSubConfig.protocol {
	case "rabbit":
		urlStr = fmt.Sprintf("%s://%s", api.pubSubConfig.protocol, target)
	case "awssqs":
		urlStr = fmt.Sprintf("%s://%s/%s?region=%s", api.pubSubConfig.protocol, api.pubSubConfig.prefix, target, api.pubSubConfig.region)
	}
	return urlStr
}
