package api

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/coreos/go-oidc"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/ourrootsorg/cms-server/service"
	"gocloud.dev/blob"

	"gocloud.dev/pubsub"

	"github.com/go-playground/validator/v10"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ourrootsorg/cms-server/model"
)

// TokenKey is the key to the token property in the request context
type TokenKey string

// TokenProperty is the name of the token property in the request context
const TokenProperty TokenKey = "token"

// UserProperty is the name of the token property in the request context
const UserProperty TokenKey = "user"

type LocalAPI interface {
	GetCategories(context.Context) (*CategoryResult, *model.Errors)
	GetCategory(ctx context.Context, id string) (*model.Category, *model.Errors)
	AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, *model.Errors)
	UpdateCategory(ctx context.Context, id string, in model.Category) (*model.Category, *model.Errors)
	DeleteCategory(ctx context.Context, id string) *model.Errors
	GetCollections(ctx context.Context /* filter/search criteria */) (*CollectionResult, *model.Errors)
	GetCollection(ctx context.Context, id string) (*model.Collection, *model.Errors)
	AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, *model.Errors)
	UpdateCollection(ctx context.Context, id string, in model.Collection) (*model.Collection, *model.Errors)
	DeleteCollection(ctx context.Context, id string) *model.Errors
	GetPosts(ctx context.Context /* filter/search criteria */) (*PostResult, *model.Errors)
	GetPost(ctx context.Context, id string) (*model.Post, *model.Errors)
	AddPost(ctx context.Context, in model.PostIn) (*model.Post, *model.Errors)
	UpdatePost(ctx context.Context, id string, in model.Post) (*model.Post, *model.Errors)
	DeletePost(ctx context.Context, id string) *model.Errors
	PostContentRequest(ctx context.Context, contentRequest ContentRequest) (*ContentResult, *model.Errors)
	GetContent(ctx context.Context, key string) ([]byte, *model.Errors)
	RetrieveUser(ctx context.Context, provider OIDCProvider, token *oidc.IDToken, rawToken string) (*model.User, *model.Errors)
	Search(ctx context.Context, req SearchRequest) (SearchResult, *model.Errors)
}

// API is the container for the apilication
type API struct {
	categoryPersister   model.CategoryPersister
	collectionPersister model.CollectionPersister
	postPersister       model.PostPersister
	userPersister       model.UserPersister
	validate            *validator.Validate
	blobStoreConfig     BlobStoreConfig
	pubSubConfig        PubSubConfig
	userCache           *lru.TwoQueueCache
	es                  *elasticsearch.Client
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
func NewAPI() (*API, error) {
	api := &API{}
	api.Validate(validator.New())
	var err error
	api.userCache, err = lru.New2Q(100)
	if err != nil {
		return nil, err
	}
	return api, nil
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

// PostPersister sets the PostPersister for the api
func (api *API) PostPersister(cp model.PostPersister) *API {
	api.postPersister = cp
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
	return pubsub.OpenTopic(ctx, api.getPubSubURLStr(topic))
}

// OpenSubscription opens a subscription to a queue
// Shutdown(ctx) the subscription when you're done with it, and ack() messages when you've processed them
func (api *API) OpenSubscription(ctx context.Context, queue string) (*pubsub.Subscription, error) {
	return pubsub.OpenSubscription(ctx, api.getPubSubURLStr(queue))
}

func (api *API) getPubSubURLStr(target string) string {
	var urlStr string
	switch api.pubSubConfig.protocol {
	case "": // use rabbit as the default protocol for testing convenience
		fallthrough
	case "rabbit":
		urlStr = fmt.Sprintf("rabbit://%s", target)
	case "awssqs":
		urlStr = fmt.Sprintf("awssqs://%s/%s?region=%s", api.pubSubConfig.prefix, target, api.pubSubConfig.region)
	default:
		log.Fatalf("Invalid protocol %s\n", api.pubSubConfig.protocol)
	}
	return urlStr
}

// UserPersister sets the UserPersister for the API
func (api *API) UserPersister(p model.UserPersister) *API {
	api.userPersister = p
	return api
}

// Elasticsearch sets the Elasticsearch client
func (api *API) Elasticsearch(es *elasticsearch.Client) *API {
	api.es = es
	return api
}
