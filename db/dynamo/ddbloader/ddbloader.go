package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/codingconcepts/env"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/logutils"
	"github.com/ourrootsorg/cms-server/persist/dynamo"
)

// placesTSV         = "https://s3.amazonaws.com/public.ourroots.org/places.tsv"
// placeWordsTSV     = "https://s3.amazonaws.com/public.ourroots.org/place_words.tsv"
// placeSettingsTSV  = "https://s3.amazonaws.com/public.ourroots.org/place_settings.tsv"

func main() {
	config, err := ParseEnv()
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR", "FATAL"},
		MinLevel: logutils.LogLevel(config.MinLogLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	cfgs := make([]*aws.Config, 0)
	if config.LocalTest {
		// Use DynamoDB local
		cfgs = append(cfgs, &aws.Config{
			Region:      aws.String(config.Region),
			Endpoint:    aws.String("http://localhost:18000"),
			DisableSSL:  aws.Bool(true),
			Credentials: credentials.NewStaticCredentials("ACCESS_KEY", "SECRET", ""),
		})
	}
	sess, err := session.NewSession(cfgs...)
	if err != nil {
		log.Fatalf("[FATAL] Error creating AWS session: %v", err)
	}
	p, err := dynamo.NewPersister(sess, config.DynamoDBTableName)
	if err != nil {
		log.Fatalf("[FATAL] Error creating DynamoDB persister: %v", err)
	}

	var fileNames string
	if config.FileURLs != "" {
		fileNames = config.FileURLs
	} else {
		fileNames = config.FilePaths
	}
	for _, fileName := range strings.Split(fileNames, ",") {
		r := openFile(config, fileName)
		defer r.Close()

		switch {
		case strings.HasSuffix(fileName, "places.tsv"):
			err = p.LoadPlaceData(r)
			if err != nil {
				log.Fatalf("[FATAL] Unable to load place data from %s: %v", fileName, err)
			}
			log.Printf("[INFO] Loaded place data from %s", fileName)
		case strings.HasSuffix(fileName, "place_settings.tsv"):
			log.Printf("[DEBUG] Loading place settings data from %s", fileName)
			err = p.LoadPlaceSettingsData(r)
			if err != nil {
				log.Fatalf("[FATAL] Unable to load place settings data from %s: %v", fileName, err)
			}
			log.Printf("[INFO] Loaded place settings data from %s", fileName)
		case strings.HasSuffix(fileName, "place_words.tsv"):
			err = p.LoadPlaceWordData(r)
			if err != nil {
				log.Fatalf("[FATAL] Unable to load place word data from %s: %v", fileName, err)
			}
			log.Printf("[INFO] Loaded place words data from %s", fileName)
		default:
			log.Fatalf("[FATAL] Don't know how to load '%s'", fileName)
		}
	}
	err = p.SetFinalThroughput()
	if err != nil {
		log.Fatalf("[FATAL] Error setting final table throughput. **WARNING** You should review the provisioned throughput ASAP, because the current throughput may be expensive!: %v", err)
	}
}

func openFile(config *Env, fileName string) io.ReadCloser {
	var reader io.ReadCloser
	if config.FileURLs != "" {
		resp, err := http.Get(fileName)
		if err != nil {
			log.Fatalf("[FATAL] Unable to open file URL %s: %v", fileName, err)
		}
		reader = resp.Body
	} else {
		f, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("[FATAL] Unable to open file path %s: %v", fileName, err)
		}
		reader = f
	}
	return reader
}

// Env holds values parse from environment variables
type Env struct {
	MinLogLevel       string `env:"MIN_LOG_LEVEL" validate:"omitempty,eq=DEBUG|eq=INFO|eq=ERROR"`
	DynamoDBTableName string `env:"DYNAMODB_TABLE_NAME" validate:"required"`
	Region            string `env:"AWS_REGION" validate:"required"`
	FileURLs          string `env:"FILE_URLS" validate:"required_without=FilePaths,omitempty"`
	FilePaths         string `env:"FILE_PATHS" validate:"required_without=FileURLs,omitempty"`
	LocalTestString   string `env:"LOCAL_TEST" validate:"omitempty,eq=true|eq=false"`
	LocalTest         bool
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
			case "LOCAL_TEST":
				errs += fmt.Sprintf("  Invalid LOCAL_TEST: '%v', valid values are 'TRUE' or 'FALSE'\n", fe.Value())
			case "AWS_REGION":
				errs += fmt.Sprintf("  AWS_REGION is required\n")
			default:
				errs += fmt.Sprintf("  Other error, fe: %#v", fe)
			}
		}
		return nil, errors.New(errs)
	}
	if config.MinLogLevel == "" {
		config.MinLogLevel = "DEBUG"
	}
	if config.FileURLs != "" && config.FilePaths != "" {
		return nil, errors.New("Must set only one of FILE_URL or FILE_PATH")
	}
	if config.LocalTestString != "" {
		config.LocalTest, err = strconv.ParseBool(config.LocalTestString)
		if err != nil {
			// should never happen
			return nil, fmt.Errorf("Couldn't parse LOCAL_TEST value '%s'", config.LocalTestString)
		}
	}
	return &config, nil
}
