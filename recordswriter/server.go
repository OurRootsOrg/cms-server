package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ourrootsorg/cms-server/model"

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
	var msg model.RecordsWriterMsg
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		log.Printf("[ERROR] Discarding unparsable message '%s': %v", string(rawMsg), err)
		return nil // Don't return an error, because parsing will never succeed
	}

	log.Printf("[DEBUG] processing %d", msg.PostID)

	// read post
	post, errs := ap.GetPost(ctx, msg.PostID)
	if errs != nil {
		log.Printf("[ERROR] Error calling GetPost on %d: %v", msg.PostID, errs)
		return errs
	}
	if post.RecordsStatus != model.PostLoading {
		log.Printf("[ERROR] post %d not Loading, is %s\n", post.ID, post.RecordsStatus)
		return nil
	}

	// delete any previous records for post
	errs = ap.DeleteRecordsForPost(ctx, post.ID)
	if errs != nil {
		log.Printf("[ERROR] DeleteRecordsForPost on %d: %v\n", post.ID, errs)
		return errs
	}

	// open bucket
	bucket, err := ap.OpenBucket(ctx)
	if err != nil {
		log.Printf("[ERROR] OpenBucket %v\n", err)
		return api.NewErrors(http.StatusInternalServerError, err)
	}
	defer bucket.Close()

	// read datas
	bs, err := bucket.ReadAll(ctx, post.RecordsKey)
	if err != nil {
		log.Printf("[ERROR] ReadAll %v\n", err)
		return api.NewErrors(http.StatusInternalServerError, err)
	}
	var datas []map[string]string
	err = json.Unmarshal(bs, &datas)
	if err != nil {
		log.Printf("[ERROR] Unmarshal datas %v\n", err)
		return api.NewErrors(http.StatusInternalServerError, err)
	}

	// set up workers
	in := make(chan map[string]string)
	out := make(chan error)
	for i := 0; i < numWorkers; i++ {
		go func(in chan map[string]string, out chan error) {
			for data := range in {
				_, errs := ap.AddRecord(ctx, model.RecordIn{
					RecordBody: model.RecordBody{
						Data: data,
					},
					Post: post.ID,
				})
				out <- errs
			}
		}(in, out)
	}

	// send datas to workers
	go func(in chan map[string]string, datas []map[string]string) {
		for _, data := range datas {
			in <- data
		}
		close(in)
	}(in, datas)

	// wait for workers to complete
	errs = nil
	for i := 0; i < len(datas); i++ {
		if e := <-out; e != nil {
			log.Printf("[ERROR] AddRecord %v\n", e)
			errs = e
		}
	}
	if errs != nil {
		return errs
	}

	// update post.recordsStatus = load complete
	post.RecordsStatus = model.PostLoadComplete
	_, errs = ap.UpdatePost(ctx, post.ID, *post)
	if errs != nil {
		log.Printf("[ERROR] UpdatePost %v\n", errs)
	}

	return errs
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
		QueueConfig("recordswriter", env.PubSubRecordsWriterURL)

	db, err := postgres.Open(ctx, env.DatabaseURL)
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
	log.Printf("Connected to %s\n", env.DatabaseURL)

	p := persist.NewPostgresPersister(db)
	ap.
		PostPersister(p).
		RecordPersister(p)
	log.Print("[INFO] Using PostgresPersister")

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
	DatabaseURL            string `env:"DATABASE_URL" validate:"required,url"`
	BaseURL                *url.URL
	Region                 string `env:"AWS_REGION"`
	BlobStoreEndpoint      string `env:"BLOB_STORE_ENDPOINT"`
	BlobStoreAccessKey     string `env:"BLOB_STORE_ACCESS_KEY"`
	BlobStoreSecretKey     string `env:"BLOB_STORE_SECRET_KEY"`
	BlobStoreBucket        string `env:"BLOB_STORE_BUCKET"`
	BlobStoreDisableSSL    bool   `env:"BLOB_STORE_DISABLE_SSL"`
	PubSubRecordsWriterURL string `env:"PUB_SUB_RECORDSWRITER_URL" validate:"required,url"`
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
			}
		}
		return nil, errors.New(errs)
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
