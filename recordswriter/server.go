package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/ourrootsorg/cms-server/stdplace"

	"github.com/ourrootsorg/cms-server/stddate"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist/dynamo"

	"github.com/ourrootsorg/cms-server/persist"
	"gocloud.dev/postgres"

	"github.com/codingconcepts/env"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/logutils"
	"github.com/ourrootsorg/cms-server/api"
)

const (
	defaultURL = "http://localhost:3000"
)

const numWorkers = 40

type workerIn struct {
	data map[string]string
	ix   int
}

type workerOut struct {
	data     map[string]string
	ix       int
	recordID uint32
	errs     error
}

type recordIndex struct {
	recordID uint32
	ix       int
}

func cleanHeader(header string) string {
	return strings.ToLower(strings.ReplaceAll(header, " ", ""))
}

func findHeader(headers []string, value string) string {
	for _, header := range headers {
		if header == value {
			return header
		}
	}
	for _, header := range headers {
		if cleanHeader(header) == cleanHeader(value) {
			return header
		}
	}
	return value
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func sameHeaders(headers1 []string, headers2 []string) bool {
	if len(headers1) != len(headers2) {
		return false
	}
	for _, header := range headers1 {
		if !containsString(headers2, header) {
			return false
		}
	}
	for _, header := range headers2 {
		if !containsString(headers1, header) {
			return false
		}
	}
	return true
}

func loadRecords(ctx context.Context, ap *api.API, post *model.Post) error {
	log.Printf("[DEBUG] Processing file %s\n", post.RecordsKey)
	// read collection for post
	collection, errs := ap.GetCollection(ctx, post.Collection)
	if errs != nil {
		log.Printf("[ERROR] GetCollection %v\n", errs)
		return errs
	}

	// identify date and place fields
	dateFields := map[string]bool{}
	placeFields := map[string]bool{}
	for _, mapping := range collection.Mappings {
		if strings.HasSuffix(mapping.IxField, "Date") {
			dateFields[mapping.Header] = true
		} else if strings.HasSuffix(mapping.IxField, "Place") {
			placeFields[mapping.Header] = true
		}
	}

	// open bucket
	bucket, err := ap.OpenBucket(ctx, false)
	if err != nil {
		log.Printf("[ERROR] OpenBucket %v\n", err)
		return api.NewError(err)
	}
	defer bucket.Close()

	// get key
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		log.Printf("[ERROR] Missing society ID %v\n", err)
		return api.NewError(err)
	}
	key := fmt.Sprintf("/%d/%s", societyID, post.RecordsKey)

	// read datas
	bs, err := bucket.ReadAll(ctx, key)
	if err != nil {
		log.Printf("[ERROR] ReadAll %v\n", err)
		return api.NewError(err)
	}

	// read CSV
	r := csv.NewReader(bytes.NewReader(bs))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		log.Printf("[ERROR] reading file: %v\n", err)
		return api.NewError(err)
	}

	// convert records into datas map
	var headers []string
	var collectionHeaders []string
	for _, collectionField := range collection.Fields {
		collectionHeaders = append(collectionHeaders, collectionField.Header)
	}
	var datas []map[string]string
	for i, record := range records {
		if i == 0 {
			// header row
			for _, field := range record {
				header := findHeader(collectionHeaders, field)
				headers = append(headers, header)
			}
			log.Printf("[DEBUG] collectionHeaders %s\n", strings.Join(collectionHeaders, ", "))
			log.Printf("[DEBUG] headers %s\n", strings.Join(headers, ", "))
			log.Printf("[DEBUG] same %t\n", sameHeaders(collectionHeaders, headers))
			if !sameHeaders(collectionHeaders, headers) {
				err := fmt.Errorf("found headers: %s; but expected headers: %s\n",
					strings.Join(headers, ", "),
					strings.Join(collectionHeaders, ", "))
				return api.NewHTTPError(err, http.StatusBadRequest)
			}
			continue
		}
		data := map[string]string{}
		for j, field := range record {
			data[headers[j]] = field
		}
		datas = append(datas, data)
	}
	log.Printf("[DEBUG] datas: %#v", datas)

	// delete any previous record households for post
	errs = ap.DeleteRecordHouseholdsForPost(ctx, post.ID)
	if errs != nil {
		log.Printf("[ERROR] DeleteRecordHouseholdsForPost on %d: %v\n", post.ID, errs)
		return errs
	}

	// delete any previous records for post
	errs = ap.DeleteRecordsForPost(ctx, post.ID)
	if errs != nil {
		log.Printf("[ERROR] DeleteRecordsForPost on %d: %v\n", post.ID, errs)
		return errs
	}

	// set up workers
	in := make(chan workerIn)
	out := make(chan workerOut)
	for i := 0; i < numWorkers; i++ {
		go func(in chan workerIn, out chan workerOut) {
			for msg := range in {
				for key := range msg.data {
					if dateFields[key] {
						var std string
						if d := stddate.Standardize(msg.data[key]); d != nil {
							std = d.Encode()
						}
						msg.data[key+stddate.StdSuffix] = std
					}
					if placeFields[key] {
						var std string
						place, err := ap.StandardizePlace(ctx, msg.data[key], collection.Location)
						if err != nil {
							log.Printf("[ERROR] Standardize place %s %v\n", msg.data[key], err)
						} else if place != nil {
							std = place.FullName
						}
						msg.data[key+stdplace.StdSuffix] = std
					}
				}

				log.Printf("[DEBUG] Processing data: %#v", msg.data)
				record, errs := ap.AddRecord(ctx, model.RecordIn{
					RecordBody: model.RecordBody{
						Data: msg.data,
					},
					Post: post.ID,
				})
				// Retry up to three times with exponential backoff of 1, 10 and 100ms
				for i, wait := 0, 1*time.Millisecond; errs != nil && i < 3; i, wait = i+1, wait*10 {
					if !isRetryable(errs) {
						log.Printf("[DEBUG] Error %#v is non-retryable", errs)
						break
					}
					time.Sleep(wait)
					log.Printf("[DEBUG] Error %#v, retry #%d %#v", errs, i+1, msg.data)
					record, errs = ap.AddRecord(ctx, model.RecordIn{
						RecordBody: model.RecordBody{
							Data: msg.data,
						},
						Post: post.ID,
					})
					if errs != nil {
						log.Printf("[DEBUG] Retry #%d failed: %#v", i+1, errs)
					}
				}
				out <- workerOut{
					data:     msg.data,
					recordID: record.ID,
					ix:       msg.ix,
					errs:     errs,
				}
			}
		}(in, out)
	}

	// send datas to workers
	go func(in chan workerIn, datas []map[string]string) {
		for ix, data := range datas {
			in <- workerIn{data: data, ix: ix}
		}
		close(in)
	}(in, datas)

	// wait for workers to complete and gather household information (if any)
	households := map[string][]recordIndex{}
	errs = nil
	for i := 0; i < len(datas); i++ {
		result := <-out
		if result.errs != nil {
			log.Printf("[ERROR] AddRecord received error: %#v", result.errs)
			errs = result.errs
			continue
		}
		if collection.HouseholdNumberHeader != "" {
			householdID := result.data[collection.HouseholdNumberHeader]
			if householdID != "" {
				households[householdID] = append(households[householdID], recordIndex{recordID: result.recordID, ix: result.ix})
			}
		}
	}
	close(out)

	// create households
	if collection.HouseholdNumberHeader != "" {
		for householdID, recordIndexes := range households {
			// sort recordIDs by index
			sort.Slice(recordIndexes, func(i, j int) bool {
				return recordIndexes[i].ix < recordIndexes[j].ix
			})
			var recordIDs []uint32
			for _, recordIndex := range recordIndexes {
				recordIDs = append(recordIDs, recordIndex.recordID)
			}
			if _, e := ap.AddRecordHousehold(ctx, model.RecordHouseholdIn{
				Post:      post.ID,
				Household: householdID,
				Records:   recordIDs,
			}); e != nil {
				log.Printf("[ERROR] AddRecordHousehold received error: %#v", e)
				errs = e
				break
			}
		}
	}

	return errs
}

func processMessage(ctx context.Context, ap *api.API, rawMsg []byte) error {
	var msg model.RecordsWriterMsg
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		log.Printf("[ERROR] Discarding unparsable message '%s': %v", string(rawMsg), err)
		return nil // Don't return an error, because parsing will never succeed
	}

	log.Printf("[DEBUG] RecordsWriter Processing PostID: %d", msg.PostID)

	sctx := utils.AddSocietyIDToContext(ctx, msg.SocietyID)

	// read post
	post, errs := ap.GetPost(sctx, msg.PostID)
	if errs != nil {
		log.Printf("[ERROR] Error calling GetPost on %d: %v", msg.PostID, errs)
		return errs
	}
	if post.RecordsStatus != model.RecordsStatusToLoad && post.RecordsStatus != model.RecordsStatusError {
		log.Printf("[ERROR] post %d not ToLoad, is %s\n", post.ID, post.RecordsStatus)
		return nil
	}

	post.RecordsStatus = model.RecordsStatusLoading
	post, errs = ap.UpdatePost(sctx, msg.PostID, *post)
	if errs != nil {
		log.Printf("[ERROR] Error calling UpdatePost on %d: %v", msg.PostID, errs)
		return errs
	}

	// do the work
	loadErrs := loadRecords(sctx, ap, post)

	// get post again, in case there were any changes in the meantime
	post, errs = ap.GetPost(sctx, msg.PostID)
	if errs != nil {
		log.Printf("[ERROR] Error calling GetPost on %d: %v", msg.PostID, errs)
		return errs
	}
	// update post
	if loadErrs != nil {
		post.RecordsStatus = model.RecordsStatusLoadError
		post.RecordsError = loadErrs.Error()
	} else {
		post.RecordsStatus = model.RecordsStatusLoadComplete
	}
	_, err = ap.UpdatePost(sctx, post.ID, *post)
	if err != nil {
		log.Printf("[ERROR] UpdatePost %v\n", err)
		return err
	}

	return errs
}

func isRetryable(err error) bool {
	er, ok := err.(*api.Error)
	return ok && er.HTTPStatus() == http.StatusConflict
}

type lambdaHandler struct {
	ap *api.API
}

func (h lambdaHandler) handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	var err error
	for _, message := range sqsEvent.Records {
		// process message
		err = processMessage(ctx, h.ap, []byte(message.Body))
		if err != nil {
			log.Printf("[ERROR] Error processing message %v", err)
			// TODO shouldn't this be break so we fail as soon as a message fails?
			continue
		}
	}
	// We'll return the last error we received, if any. That will fail the batch, which will be retried.
	return err
}

func main() {
	ctx := context.Background()

	// parse environment
	env, err := ParseEnv()
	if err != nil {
		log.Fatalf("[FATAL] Error parsing environment variables: %v", err)
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR", "FATAL"},
		MinLevel: logutils.LogLevel(env.MinLogLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	// configure api
	ap, err := api.NewAPI()
	if err != nil {
		log.Fatalf("Error calling NewAPI: %v", err)
	}
	defer ap.Close()
	ap = ap.
		BlobStoreConfig(env.Region, env.BlobStoreEndpoint, env.BlobStoreAccessKey, env.BlobStoreSecretKey, env.BlobStoreBucket, env.BlobStoreDisableSSL).
		QueueConfig("recordswriter", env.PubSubRecordsWriterURL).
		QueueConfig("imageswriter", env.PubSubImagesWriterURL)

	if env.DatabaseURL != "" {
		// Don't leak credentials from URL
		dbURL, err := url.Parse(env.DatabaseURL)
		if err != nil {
			log.Fatalf("[FATAL] Bad database URL: %v", err)
		}
		log.Printf("[INFO] Connecting to %s\n", dbURL.Host)
		db, err := postgres.Open(context.TODO(), env.DatabaseURL)
		defer db.Close()
		if err != nil {
			log.Fatalf("[FATAL] Error opening database connection: %v\n  DATABASE_URL: %s",
				err,
				env.DatabaseURL,
			)
		}

		// ping the database to make sure we can connect
		cnt := 0
		err = errors.New("unknown error")
		for err != nil && cnt <= 3 {
			if cnt > 0 {
				time.Sleep(time.Duration(math.Pow(2.0, float64(cnt))) * time.Second)
			}
			err = db.Ping()
			cnt++
		}
		if err != nil {
			log.Fatalf("[FATAL] Error connecting to database: %v\n DATABASE_URL: %s\n",
				err,
				env.DatabaseURL,
			)
		}
		log.Printf("[INFO] Connected to %s\n", dbURL.Host)
		p := persist.NewPostgresPersister(db)
		ap.
			CollectionPersister(p).
			PostPersister(p).
			RecordPersister(p).
			PlaceStandardizer(ctx, p)
		if err != nil {
			log.Fatalf("[FATAL] Error initializing place standardizer %v\n", err)
		}
		log.Print("[INFO] Using PostgresPersister")
	} else {
		sess, err := session.NewSession()
		if err != nil {
			log.Fatalf("[FATAL] Error creating AWS session: %v", err)
		}
		p, err := dynamo.NewPersister(sess, env.DynamoDBTableName)
		if err != nil {
			log.Fatalf("[FATAL] Error creating DynamoDB persister: %v", err)
		}
		ap.
			CollectionPersister(p).
			PostPersister(p).
			RecordPersister(p).
			PlaceStandardizer(ctx, p)
		// This doesn't do anything
		// if err != nil {
		// 	log.Fatalf("[FATAL] Error initializing place standardizer %v\n", err)
		// }
		log.Print("[INFO] Using DynamoDBPersister")
	}

	if env.IsLambda {
		h := lambdaHandler{ap: ap}
		lambda.Start(h.handler)
	} else {
		// subscribe to recordswriter queue
		sub, err := ap.OpenSubscription(ctx, "recordswriter")
		if err != nil {
			log.Fatalf("[FATAL] Can't open subscription %v\n", err)
		}
		defer sub.Shutdown(ctx)

		for {
			msg, err := sub.Receive(ctx)
			if err != nil {
				log.Printf("[ERROR] Receiving message %v\n", err)
				continue
			}
			// process message
			errs := processMessage(ctx, ap, msg.Body)
			if errs != nil {
				log.Printf("[ERROR] Processing message %v\n", errs)
				continue
			}
			msg.Ack()
		}
	}
}

// Env holds values parse from environment variables
type Env struct {
	LambdaTaskRoot         string `env:"LAMBDA_TASK_ROOT"`
	IsLambda               bool
	MinLogLevel            string `env:"MIN_LOG_LEVEL" validate:"omitempty,eq=DEBUG|eq=INFO|eq=ERROR"`
	BaseURLString          string `env:"BASE_URL" validate:"omitempty,url"`
	DatabaseURL            string `env:"DATABASE_URL" validate:"required_without=DynamoDBTableName,omitempty,url"`
	DynamoDBTableName      string `env:"DYNAMODB_TABLE_NAME" validate:"required_without=DatabaseURL"`
	BaseURL                *url.URL
	Region                 string `env:"AWS_REGION"`
	BlobStoreEndpoint      string `env:"BLOB_STORE_ENDPOINT"`
	BlobStoreAccessKey     string `env:"BLOB_STORE_ACCESS_KEY"`
	BlobStoreSecretKey     string `env:"BLOB_STORE_SECRET_KEY"`
	BlobStoreBucket        string `env:"BLOB_STORE_BUCKET"`
	BlobStoreDisableSSL    bool   `env:"BLOB_STORE_DISABLE_SSL"`
	PubSubRecordsWriterURL string `env:"PUB_SUB_RECORDSWRITER_URL" validate:"required,url"`
	PubSubImagesWriterURL  string `env:"PUB_SUB_IMAGESWRITER_URL" validate:"required,url"`
}

// ParseEnv parses and validates environment variables and stores them in the Env structure
func ParseEnv() (*Env, error) {
	var config Env
	if err := env.Set(&config); err != nil {
		log.Fatal(err)
	}
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("env")
	})
	err := validate.Struct(config)
	if err != nil {
		errs := "Error parsing environment variables:\n"
		for _, fe := range err.(validator.ValidationErrors) {
			switch fe.Field() {
			case "MIN_LOG_LEVEL":
				errs += fmt.Sprintf("  Invalid MIN_LOG_LEVEL: '%v', valid values are 'DEBUG', 'INFO' or 'ERROR'\n", fe.Value())
			case "BASE_URL":
				errs += fmt.Sprintf("  Invalid BASE_URL: '%v' is not a valid URL\n", fe.Value())
			case "DATABASE_URL":
				errs += fmt.Sprintf("  Invalid DATABASE_URL: '%v' is not a valid PostgreSQL URL\n", fe.Value())
			case "PUB_SUB_RECORDSWRITER_URL":
				errs += fmt.Sprintf("  Invalid PUB_SUB_RECORDSWRITER_URL: '%v'is not a valid URL\n", fe.Value())
			case "PUB_SUB_IMAGESWRITER_URL":
				errs += fmt.Sprintf("  Invalid PUB_SUB_IMAGESWRITER_URL: '%v'is not a valid URL\n", fe.Value())
			}
		}
		return nil, errors.New(errs)
	}
	if config.DatabaseURL != "" && config.DynamoDBTableName != "" {
		return nil, errors.New("Must only set one of DATABASE_URL or DYNAMODB_TABLE_NAME")
	}
	config.IsLambda = config.LambdaTaskRoot != ""
	if config.MinLogLevel == "" {
		config.MinLogLevel = "DEBUG"
	}
	if config.BaseURLString == "" {
		config.BaseURLString = defaultURL
	}
	config.BaseURL, err = url.ParseRequestURI(config.BaseURLString)
	if err != nil {
		// Unreachable, if the validator does its job
		return nil, fmt.Errorf("unable to parse BASE_URL '%s': %v", config.BaseURLString, err)
	}
	return &config, nil
}
