package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/ourrootsorg/cms-server/stdplace"

	"github.com/cenkalti/backoff/v4"

	"github.com/ourrootsorg/go-oidc"
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

// LocalAPI is an interface used for mocking API
type LocalAPI interface {
	GetCategories(context.Context) (*CategoryResult, error)
	GetCategoriesByID(ctx context.Context, ids []uint32) ([]model.Category, error)
	GetCategory(ctx context.Context, id uint32) (*model.Category, error)
	AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, error)
	UpdateCategory(ctx context.Context, id uint32, in model.Category) (*model.Category, error)
	DeleteCategory(ctx context.Context, id uint32) error
	GetCollections(ctx context.Context /* filter/search criteria */) (*CollectionResult, error)
	GetCollectionsByID(ctx context.Context, ids []uint32) ([]model.Collection, error)
	GetCollection(ctx context.Context, id uint32) (*model.Collection, error)
	AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, error)
	UpdateCollection(ctx context.Context, id uint32, in model.Collection) (*model.Collection, error)
	DeleteCollection(ctx context.Context, id uint32) error
	GetPosts(ctx context.Context /* filter/search criteria */) (*PostResult, error)
	GetPost(ctx context.Context, id uint32) (*model.Post, error)
	GetPostImage(ctx context.Context, id uint32, filePath string, expireSeconds, height, width int) (*ImageMetadata, error)
	AddPost(ctx context.Context, in model.PostIn) (*model.Post, error)
	UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, error)
	DeletePost(ctx context.Context, id uint32) error
	PostContentRequest(ctx context.Context, contentRequest ContentRequest) (*ContentResult, error)
	GetContent(ctx context.Context, key string) ([]byte, error)
	RetrieveUser(ctx context.Context, provider OIDCProvider, token *oidc.IDToken, rawToken string) (*model.User, error)
	GetRecordsForPost(ctx context.Context, postid uint32) (*RecordResult, error)
	GetRecordsByID(ctx context.Context, ids []uint32) ([]model.Record, error)
	GetRecord(ctx context.Context, id uint32) (*model.Record, error)
	AddRecord(ctx context.Context, in model.RecordIn) (*model.Record, error)
	UpdateRecord(ctx context.Context, id uint32, in model.Record) (*model.Record, error)
	DeleteRecord(ctx context.Context, id uint32) error
	DeleteRecordsForPost(ctx context.Context, postID uint32) error
	GetRecordHouseholdsForPost(ctx context.Context, postid uint32) ([]model.RecordHousehold, error)
	GetRecordHousehold(ctx context.Context, postID uint32, householdID string) (*model.RecordHousehold, error)
	AddRecordHousehold(ctx context.Context, in model.RecordHouseholdIn) (*model.RecordHousehold, error)
	DeleteRecordHouseholdsForPost(ctx context.Context, postID uint32) error
	Search(ctx context.Context, req *SearchRequest) (*model.SearchResult, error)
	SearchByID(ctx context.Context, id string) (*model.SearchHit, error)
	SearchDeleteByID(ctx context.Context, id string) error
	GetSettings(ctx context.Context) (*model.Settings, error)
	UpdateSettings(ctx context.Context, in model.Settings) (*model.Settings, error)
	StandardizePlace(ctx context.Context, text, defaultContainingPlace string) (*model.Place, error)
	GetPlacesByPrefix(ctx context.Context, prefix string, count int) ([]model.Place, error)
	GetNameVariants(ctx context.Context, nameType model.NameType, name string) (*model.NameVariants, error)
}

// API is the container for the apilication
type API struct {
	categoryPersister        model.CategoryPersister
	collectionPersister      model.CollectionPersister
	postPersister            model.PostPersister
	recordPersister          model.RecordPersister
	userPersister            model.UserPersister
	placePersister           model.PlacePersister
	namePersister            model.NamePersister
	settingsPersister        model.SettingsPersister
	validate                 *validator.Validate
	blobStoreConfig          BlobStoreConfig
	pubSubConfig             PubSubConfig
	placeStandardizer        *stdplace.Standardizer
	userCache                *lru.TwoQueueCache
	nameVariantsCache        *lru.TwoQueueCache
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
	api.nameVariantsCache, err = lru.New2Q(100)
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
	if api.placeStandardizer != nil {
		api.placeStandardizer.Close()
		api.placeStandardizer = nil
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

// QueueConfig configures queues
func (api *API) QueueConfig(queueName, queueURL string) *API {
	api.pubSubConfig.queueURL[queueName] = queueURL
	return api
}

// UserPersister sets the UserPersister for the API
func (api *API) UserPersister(p model.UserPersister) *API {
	api.userPersister = p
	return api
}

// PlacePersister sets the PostPersister for the api
func (api *API) PlacePersister(p model.PlacePersister) *API {
	api.placePersister = p
	return api
}

// PlaceStandardizer sets the placeStandardizer for the api
func (api *API) PlaceStandardizer(ctx context.Context, p model.PlacePersister) *API {
	std, err := stdplace.NewStandardizer(ctx, p)
	if err != nil {
		log.Fatalf("[FATAL] Error initializing place standardizer %v\n", err)
	}
	api.placeStandardizer = std
	return api
}

// NamePersister sets the NamePersister for the api
func (api *API) NamePersister(p model.NamePersister) *API {
	api.namePersister = p
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

// Error is an ordered collection of errors
type Error struct {
	errs       []model.Error
	httpStatus int
}

// fromError builds an Errors collection from a `*model.Error`
func fromError(e *model.Error) *Error {
	var httpStatus int
	switch e.Code {
	case model.ErrBadReference:
		httpStatus = http.StatusBadRequest
	case model.ErrConcurrentUpdate:
		httpStatus = http.StatusConflict
	case model.ErrConflict:
		httpStatus = http.StatusConflict
	case model.ErrNotFound:
		httpStatus = http.StatusNotFound
	case model.ErrRequired:
		httpStatus = http.StatusBadRequest
	case model.ErrOther:
		httpStatus = http.StatusInternalServerError
	default: // Shouldn't hit this unless someone adds a new code
		log.Printf("[INFO] Encountered unexpected error code: %s", e.Code)
		httpStatus = http.StatusInternalServerError
	}

	return &Error{
		errs:       []model.Error{*e},
		httpStatus: httpStatus,
	}
}

// NewHTTPError allows overriding the HTTP status code inferred by `NewErrror`.
func NewHTTPError(err error, httpStatus int) *Error {
	e := NewError(err)
	e.httpStatus = httpStatus
	return e
}

// NewError builds an Error collection from an `error`, which may actually be a ValidationErrors collection
// or a `model.Error`
func NewError(err error) *Error {
	var e *model.Error
	var isModelError bool
	e, isModelError = err.(*model.Error)
	if !isModelError {
		e1, ok := err.(model.Error)
		if ok {
			e = &e1
			isModelError = true
		}
	}
	// if httpStatus <= 0 && !isModelError {
	// 	log.Printf("[INFO] Warning httpStatus = %d and err is not a model.Error: %#v", httpStatus, err)
	// 	httpStatus = http.StatusInternalServerError
	// }
	if isModelError {
		// Note that this ignores `httpStatus`
		return fromError(e)
	}

	errors := Error{
		errs: make([]model.Error, 0),
		// httpStatus: httpStatus,
	}

	if ves, ok := err.(validator.ValidationErrors); ok {
		errors.httpStatus = http.StatusBadRequest
		for _, fe := range ves {
			if fe.Tag() == "required" {
				name := strings.SplitN(fe.Namespace(), ".", 2)
				// log.Printf("name: %v", name)
				errors.errs = append(errors.errs, *model.NewError(model.ErrRequired, name[1]))
			} else {
				errors.errs = append(errors.errs, *model.NewError(model.ErrOther, fmt.Sprintf("Key: '%s' Error: Field validation for '%s' failed on the '%s' tag", fe.Namespace(), fe.Field(), fe.Tag())))
			}
		}
	} else {
		errors.httpStatus = http.StatusInternalServerError
		errors.errs = append(errors.errs, *model.NewError(model.ErrOther, err.Error()))
	}
	return &errors
}

// HTTPStatus returns the HTTP status code
func (e Error) HTTPStatus() int {
	return e.httpStatus
}

// Errs returns the slice of model.Error structs
func (e Error) Errs() []model.Error {
	return e.errs
}

func (e Error) Error() string {
	s := "Errors:"
	for _, er := range e.Errs() {
		s += "\n  " + er.Error()
	}
	return s
}
