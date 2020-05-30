package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"reflect"
	"time"

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

func processMessage(ctx context.Context, ap *api.API, msg api.RecordsWriterMsg) error {
	// TODO read post
	// TODO read blob
	// TODO write records
	// TODO update post.recordsStatus = READY
	return nil
}

func main() {
	ctx := context.Background()

	// parse environment
	env, err := ParseEnv()
	if err != nil {
		log.Fatalf("[FATAL] Error parsing environmet variables: %v", err)
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
	ap = ap.
		BlobStoreConfig(env.Region, env.BlobStoreEndpoint, env.BlobStoreAccessKey, env.BlobStoreSecretKey, env.BlobStoreBucket, env.BlobStoreDisableSSL).
		PubSubConfig(env.Region, env.PubSubProtocol, env.PubSubPrefix)
	db, err := postgres.Open(context.TODO(), env.DatabaseURL)
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

	p := persist.NewPostgresPersister(env.BaseURL.Path, db)
	ap.
		PostPersister(p).
		RecordPersister(p)
	log.Print("[INFO] Using PostgresPersister")

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
			break
		}
		var rwMsg api.RecordsWriterMsg
		err = json.Unmarshal(msg.Body, &rwMsg)
		if err != nil {
			log.Printf("[ERROR] Unmarshalling message %v\n", err)
			break
		}
		// process message
		err = processMessage(ctx, ap, rwMsg)
		if err != nil {
			log.Printf("[ERROR] Processing message %v\n", err)
			break
		}
		msg.Ack()
	}
}

// Env holds values parse from environment variables
type Env struct {
	MinLogLevel         string `env:"MIN_LOG_LEVEL" validate:"omitempty,eq=DEBUG|eq=INFO|eq=ERROR"`
	BaseURLString       string `env:"BASE_URL" validate:"omitempty,url"`
	DatabaseURL         string `env:"DATABASE_URL" validate:"required,url"`
	BaseURL             *url.URL
	Region              string `env:"AWS_REGION"`
	BlobStoreEndpoint   string `env:"BLOB_STORE_ENDPOINT"`
	BlobStoreAccessKey  string `env:"BLOB_STORE_ACCESS_KEY"`
	BlobStoreSecretKey  string `env:"BLOB_STORE_SECRET_KEY"`
	BlobStoreBucket     string `env:"BLOB_STORE_BUCKET"`
	BlobStoreDisableSSL bool   `env:"BLOB_STORE_DISABLE_SSL"`
	PubSubProtocol      string `env:"PUB_SUB_PROTOCOL"`
	PubSubPrefix        string `env:"PUB_SUB_PREFIX"`
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
			}
		}
		return nil, errors.New(errs)
	}
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
