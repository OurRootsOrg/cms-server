package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

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

const numWorkers = 10

func processMessage(ctx context.Context, ap *api.API, rawMsg []byte) error {
	var msg model.ImagesWriterMsg
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		log.Printf("[ERROR] Discarding unparsable message '%s': %v", string(rawMsg), err)
		return nil // Don't return an error, because parsing will never succeed
	}

	log.Printf("[DEBUG] Processing PostID: %d", msg.PostID)

	// read post
	post, errs := ap.GetPost(ctx, msg.PostID)
	if errs != nil {
		log.Printf("[ERROR] Error calling GetPost on %d: %v", msg.PostID, errs)
		return errs
	}
	if post.ImagesStatus != model.PostLoading {
		log.Printf("[ERROR] post %d not Loading, is %s\n", post.ID, post.ImagesStatus)
		return nil
	}

	// open bucket
	bucket, err := ap.OpenBucket(ctx)
	if err != nil {
		log.Printf("[ERROR] OpenBucket %v\n", err)
		return api.NewError(err)
	}
	defer bucket.Close()
	// set up workers
	in := make(chan *zip.File)
	out := make(chan error)
	for i := 0; i < numWorkers; i++ {
		go func(in chan *zip.File, out chan error) {
			for f := range in {
				errs = nil
				if f.UncompressedSize == 0 {
					// skip directories and empty files
					out <- errs
					continue
				}
				log.Printf("[DEBUG] Processing file: %s", f.Name)
				rc, errs := f.Open()
				if errs != nil {
					log.Printf("[ERROR] Error opening %s: %v", f.Name, errs)
					out <- errs
					continue
				}
				defer rc.Close()
				fileBytes, errs := ioutil.ReadAll(rc)
				if errs != nil && errs != io.EOF {
					log.Printf("[ERROR] Error reading contents of %s: %v", f.Name, errs)
					out <- errs
					continue
				}
				contentType := http.DetectContentType(fileBytes)
				if !strings.HasPrefix(contentType, "image") {
					log.Printf("[INFO] Skipping file %s, content type %s, because it's not an image.", f.Name, contentType)
					out <- errs
					continue
				}
				name := fmt.Sprintf(api.ImagesPrefix, msg.PostID) + f.Name
				errs = bucket.WriteAll(ctx, name, fileBytes, nil)
				out <- errs
			}
		}(in, out)
	}
	var fileCount int
	for _, zipName := range msg.NewZips {
		// Read zip file data
		ra, err := NewBucketReaderAt(ctx, bucket, zipName)
		if err != nil {
			log.Printf("[ERROR] Error opening zip file %s: %v\n", zipName, err)
			return api.NewError(err)
		}
		zr, err := zip.NewReader(ra, ra.Size)
		if err != nil {
			log.Printf("[ERROR] Error reading zip file %s: %v\n", zipName, err)
			return api.NewError(err)
		}

		// send files to workers
		go func(in chan *zip.File, files []*zip.File) {
			for _, f := range files {
				in <- f
			}
		}(in, zr.File)
		fileCount += len(zr.File)
	}
	// wait for workers to complete
	errs = nil
	for i := 0; i < fileCount; i++ {
		if e := <-out; e != nil {
			log.Printf("[ERROR] Error saving image file: %#v", e)
			errs = e
		}
	}
	close(in)
	close(out)

	// TODO we need a better way to notify the user of errors; this doesn't tell the user that anything went wrong
	if errs != nil {
		post.ImagesStatus = model.PostDraft
	} else {
		post.ImagesStatus = model.PostLoadComplete
	}
	_, err = ap.UpdatePost(ctx, post.ID, *post)
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
			PostPersister(p)
			// ImagePersister(p)
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
			PostPersister(p)
			// ImagePersister(p)
		log.Print("[INFO] Using DynamoDBPersister")
	}

	if env.IsLambda {
		h := lambdaHandler{ap: ap}
		lambda.Start(h.handler)
	} else {
		// subscribe to imageswriter queue
		sub, err := ap.OpenSubscription(ctx, "imageswriter")
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
			log.Printf("[DEBUG] Received message '%s'", string(msg.Body))
			// process message
			errs := processMessage(ctx, ap, msg.Body)
			if errs != nil {
				log.Printf("[ERROR] Processing message %v\n", errs)
				continue
			}
			log.Printf("[DEBUG] Processed message '%s'", string(msg.Body))
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
