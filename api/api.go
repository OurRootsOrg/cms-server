package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/coreos/go-oidc"
	"github.com/streadway/amqp"

	"github.com/elastic/go-elasticsearch/v7"

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
	GetCategoriesByID(ctx context.Context, ids []uint32) ([]model.Category, *model.Errors)
	GetCategory(ctx context.Context, id uint32) (*model.Category, *model.Errors)
	AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, *model.Errors)
	UpdateCategory(ctx context.Context, id uint32, in model.Category) (*model.Category, *model.Errors)
	DeleteCategory(ctx context.Context, id uint32) *model.Errors
	GetCollections(ctx context.Context /* filter/search criteria */) (*CollectionResult, *model.Errors)
	GetCollectionsByID(ctx context.Context, ids []uint32) ([]model.Collection, *model.Errors)
	GetCollection(ctx context.Context, id uint32) (*model.Collection, *model.Errors)
	AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, *model.Errors)
	UpdateCollection(ctx context.Context, id uint32, in model.Collection) (*model.Collection, *model.Errors)
	DeleteCollection(ctx context.Context, id uint32) *model.Errors
	GetPosts(ctx context.Context /* filter/search criteria */) (*PostResult, *model.Errors)
	GetPost(ctx context.Context, id uint32) (*model.Post, *model.Errors)
	AddPost(ctx context.Context, in model.PostIn) (*model.Post, *model.Errors)
	UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, *model.Errors)
	DeletePost(ctx context.Context, id uint32) *model.Errors
	PostContentRequest(ctx context.Context, contentRequest ContentRequest) (*ContentResult, *model.Errors)
	GetContent(ctx context.Context, key string) ([]byte, *model.Errors)
	RetrieveUser(ctx context.Context, provider OIDCProvider, token *oidc.IDToken, rawToken string) (*model.User, *model.Errors)
	GetRecordsForPost(ctx context.Context, postid uint32) (*RecordResult, *model.Errors)
	Search(ctx context.Context, req *SearchRequest) (*model.SearchResult, *model.Errors)
	SearchByID(ctx context.Context, id string) (*model.SearchHit, *model.Errors)
	SearchDeleteByID(ctx context.Context, id string) *model.Errors
	GetSettings(ctx context.Context) (*model.Settings, *model.Errors)
	UpdateSettings(ctx context.Context, in model.Settings) (*model.Settings, *model.Errors)
}

// API is the container for the apilication
type API struct {
	categoryPersister        model.CategoryPersister
	collectionPersister      model.CollectionPersister
	postPersister            model.PostPersister
	recordPersister          model.RecordPersister
	userPersister            model.UserPersister
	settingsPersister        model.SettingsPersister
	validate                 *validator.Validate
	blobStoreConfig          BlobStoreConfig
	pubSubConfig             PubSubConfig
	userCache                *lru.TwoQueueCache
	rabbitmqTopicConn        *amqp.Connection
	rabbitmqSubscriptionConn *amqp.Connection
	es                       *elasticsearch.Client
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
	queueURL map[string]string
}

// QueueURL returns the URL for a queue
func (c PubSubConfig) QueueURL(queueName string) string {
	return c.queueURL[queueName]
}

// NewAPI builds an API; Close() the api when you're done with it to free up resources
func NewAPI() (*API, error) {
	api := &API{}
	api.Validate(validator.New())
	var err error
	api.userCache, err = lru.New2Q(100)
	if err != nil {
		return nil, err
	}
	api.pubSubConfig = PubSubConfig{queueURL: map[string]string{}}
	return api, nil
}

// Close frees up any held resources
func (api *API) Close() error {
	var err error
	if api.rabbitmqTopicConn != nil {
		if e := api.rabbitmqTopicConn.Close(); e != nil {
			err = e
		}
		api.rabbitmqTopicConn = nil
	}
	if api.rabbitmqSubscriptionConn != nil {
		if e := api.rabbitmqSubscriptionConn.Close(); e != nil {
			err = e
		}
		api.rabbitmqSubscriptionConn = nil
	}
	return err
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

// RecordPersister sets the RecordPersister for the api
func (api *API) RecordPersister(cp model.RecordPersister) *API {
	api.recordPersister = cp
	return api
}

// SettingsPersister sets the SettingsPersister for the api
func (api *API) SettingsPersister(cp model.SettingsPersister) *API {
	api.settingsPersister = cp
	return api
}

// BlobStoreConfig configures the blob store service
func (api *API) BlobStoreConfig(region, endpoint, accessKeyID, secretAccessKey, bucket string, disableSSL bool) *API {
	api.blobStoreConfig = BlobStoreConfig{region, endpoint, accessKeyID, secretAccessKey, bucket, disableSSL}
	return api
}

// QueueConfig configures the recordswriter queue
func (api *API) QueueConfig(queueName, queueURL string) *API {
	api.pubSubConfig.queueURL[queueName] = queueURL
	return api
}

// UserPersister sets the UserPersister for the API
func (api *API) UserPersister(p model.UserPersister) *API {
	api.userPersister = p
	return api
}

// ElasticsearchConfig sets the Elasticsearch client
func (api *API) ElasticsearchConfig(esURL string, transport http.RoundTripper) *API {
	retryBackoff := backoff.NewExponentialBackOff()

	log.Printf("Connecting to %s\n", esURL)
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esURL},
		// Retry on 429 TooManyRequests statuses
		RetryOnStatus: []int{502, 503, 504, 429},
		// Configure the backoff function
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		// Retry up to 5 attempts
		MaxRetries: 5,
		Transport:  transport,
	})
	if err != nil {
		log.Fatalf("[FATAL] Error opening elasticsearch connection: %v\n  ELASTICSEARCH_URL: %s", err, esURL)
	}

	// ping elasticsearch to make sure we can connect
	cnt := 0
	err = errors.New("unknown error")
	for err != nil && cnt <= 6 {
		if cnt > 0 {
			time.Sleep(time.Duration(math.Pow(2.0, float64(cnt))) * time.Second)
		}
		err = pingElasticsearch(es)
		if err != nil {
			log.Printf("Elasticsearch connection error %v", err)
		}
		cnt++
	}
	if err != nil {
		log.Fatalf("[FATAL] Error connecting to elasticsearch: %v\n ELASTICSEARCH_URL: %s\n", err, esURL)
	}
	log.Printf("Connected to elasticsearch %s\n", esURL)

	api.es = es
	return api
}

func pingElasticsearch(es *elasticsearch.Client) error {
	var r map[string]interface{}

	res, err := es.Info()
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return errors.New(res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return err
	}
	// Print client and server version numbers.
	log.Printf("[DEBUG] Elasticsearch client: %s", elasticsearch.Version)
	log.Printf("[DEBUG] Elasticsearch server: %s", r["version"].(map[string]interface{})["number"])
	return nil
}

// func checkErr(err error) *model.Errors {
// 	e, ok := err.(model.Error)
// 	if ok {
// 		return model.NewErrorsFromError(e)
// 	}
// 	return model.NewErrors(http.StatusInternalServerError, err)
// }
