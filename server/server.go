//go:generate sh -c "swag init --dir $(dirname $(pwd)) --output ./docs --generalInfo server/$GOFILE"
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/codingconcepts/env"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/handlers"
	"github.com/hashicorp/logutils"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/server/docs"
	"gocloud.dev/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	awsgosignv4 "github.com/jriquelme/awsgosigv4"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	defaultURL = "http://localhost:3000"
)

// @title OurRoots API
// @version 0.1.0
// @description This is the OurRoots API

// @contact.name Jim Ancona
// @contact.url https://github.com/jancona
// @contact.email jim@anconafamily.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.ourroots.org
// @BasePath /
// @accept application/json
// @produce application/json
// @schemes http https

// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationurl https://ourroots.auth0.com/authorize?audience=https%3A%2F%2Fapi.ourroots.org%2Fpreprod
// @scope.cms Grants read and write access to the CMS
// @scope.openid Indicates that the application intends to use OIDC to verify the user's identity
// @scope.profile Grants access to OIDC user profile attributes
// @scope.email Grants access to OIDC email attributes

func main() {
	progdir := filepath.Dir(os.Args[0]) + "/"
	log.Printf("program: %s, dir: %s", os.Args[0], progdir)
	env, err := ParseEnv()
	if err != nil {
		log.Fatalf("[FATAL] Error parsing environmet variables: %v", err)
	}
	log.Printf("[DEBUG] oidcAudience: %#v, oidcDomain: %#v", env.OIDCAudience, env.OIDCDomain)

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR", "FATAL"},
		MinLevel: logutils.LogLevel(env.MinLogLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	log.Printf("[INFO] env.BaseURLString: %s, env.BaseURL.Path: %s", env.BaseURLString, env.BaseURL.Path)

	model.Initialize(env.BaseURL.Path)
	ap, err := api.NewAPI()
	if err != nil {
		log.Fatalf("Error calling NewAPI: %v", err)
	}
	var esTransport http.RoundTripper
	if env.IsLambda {
		credentials := credentials.NewEnvCredentials()
		esTransport = &awsgosignv4.SignV4SDKV1{
			RoundTripper: http.DefaultTransport,
			Credentials:  credentials,
			Region:       env.Region,
			Service:      "es",
			Now:          time.Now,
		}
	}
	defer ap.Close()
	ap = ap.
		BlobStoreConfig(env.Region, env.BlobStoreEndpoint, env.BlobStoreAccessKey, env.BlobStoreSecretKey, env.BlobStoreBucket, env.BlobStoreDisableSSL).
		QueueConfig("publisher", env.PubSubPublisherURL).
		QueueConfig("recordswriter", env.PubSubRecordsWriterURL).
		ElasticsearchConfig(env.ElasticsearchURLString, esTransport)
	app := NewApp().BaseURL(*env.BaseURL).API(ap).OIDC(env.OIDCAudience, env.OIDCDomain)
	if env.BaseURL.Scheme == "https" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
	// Only migrate if MIGRATION_DATABASE_URL is set
	if env.MigrationDatabaseURL != "" {
		func() {
			// Do database migrations
			log.Printf("[DEBUG] env.MigrationDatabaseURL = '%s'", env.MigrationDatabaseURL)
			log.Printf("[INFO] Performing migrations, if necessary")
			migrator, err := migrate.New("file://"+progdir+"../db/migrations", env.MigrationDatabaseURL)
			if err != nil {
				log.Fatalf("[FATAL] Error creating database migrator: %v", err)
			}
			defer migrator.Close()
			err = migrator.Up()
			if err == migrate.ErrNoChange {
				log.Print("[INFO] No migrations to perform")
			} else if err != nil {
				log.Fatalf("[FATAL] Error migrating database: %v", err)
			} else {
				log.Print("[INFO] Finished migrations")
			}
		}()
	}
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
	p := persist.NewPostgresPersister(env.BaseURL.Path, db)
	ap.
		CategoryPersister(p).
		CollectionPersister(p).
		PostPersister(p).
		RecordPersister(p).
		UserPersister(p)
	log.Print("[INFO] Using PostgresPersister")
	r := app.NewRouter()
	docs.SwaggerInfo.Host = env.BaseURL.Hostname()
	if env.BaseURL.Port() != "" {
		docs.SwaggerInfo.Host += ":" + env.BaseURL.Port()
	}
	docs.SwaggerInfo.BasePath = env.BaseURL.Path
	r.PathPrefix(env.BaseURL.Path + "/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(env.BaseURLString+"/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))
	r.NotFoundHandler = http.HandlerFunc(NotFound)

	if env.IsLambda {
		// Lambda-specific setup
		// Note that the Lamda doesn't serve static content, only the API
		// API Gateway proxies static content requests directly to an S3 bucket
		adapter := gorillamux.New(r)
		lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			log.Printf("[DEBUG] Lambda request %#v", req)
			// If no name is provided in the HTTP request body, throw an error
			return adapter.ProxyWithContext(ctx, req)
		})
		log.Fatal("Lambda exiting...")
	} else {
		// If we're not running in Lambda we also serve the static content.
		// This is useful in development. It might also be in a traditional server deploy, but requirements
		// for all of this are TBD.
		// uiDir := "../ui/build/web"
		// r.PathPrefix("/flutter/").
		// 	Handler(http.StripPrefix("/flutter", http.FileServer(http.Dir(flutterDir))))
		// r.PathPrefix("/wasm/").
		// 	Handler(http.StripPrefix("/wasm", http.FileServer(http.Dir(vectyDir))))
		if env.BaseURL.Scheme == "https" {
			log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%s", env.BaseURL.Port()),
				"server.crt", "server.key",
				handlers.LoggingHandler(
					os.Stdout,
					handlers.CORS(
						handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
						handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}),
						handlers.AllowedOrigins([]string{"*"}))(r)),
			))
		} else {
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.BaseURL.Port()),
				handlers.LoggingHandler(os.Stdout,
					handlers.CORS(
						handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
						handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}),
						handlers.AllowedOrigins([]string{"*"}))(r)),
			))
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
	MigrationDatabaseURL   string `env:"MIGRATION_DATABASE_URL" validate:"omitempty,url"`
	BaseURL                *url.URL
	Region                 string `env:"AWS_REGION"`
	BlobStoreEndpoint      string `env:"BLOB_STORE_ENDPOINT"`
	BlobStoreAccessKey     string `env:"BLOB_STORE_ACCESS_KEY"`
	BlobStoreSecretKey     string `env:"BLOB_STORE_SECRET_KEY"`
	BlobStoreBucket        string `env:"BLOB_STORE_BUCKET"`
	BlobStoreDisableSSL    bool   `env:"BLOB_STORE_DISABLE_SSL"`
	PubSubRecordsWriterURL string `env:"PUB_SUB_RECORDSWRITER_URL" validate:"required,url"`
	PubSubPublisherURL     string `env:"PUB_SUB_PUBLISHER_URL" validate:"required,url"`
	OIDCAudience           string `env:"OIDC_AUDIENCE" validate:"omitempty"`
	OIDCDomain             string `env:"OIDC_DOMAIN" validate:"omitempty"`
	ElasticsearchURLString string `env:"ELASTICSEARCH_URL" validate:"required,url"`
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
			case "MIGRATION_DATABASE_URL":
				errs += fmt.Sprintf("  Invalid MIGRATION_DATABASE_URL: '%v' is not a valid PostgreSQL URL\n", fe.Value())
			case "PUB_SUB_RECORDSWRITER_URL":
				errs += fmt.Sprintf("  Invalid PUB_SUB_RECORDSWRITER_URL: '%v'is not a valid URL\n", fe.Value())
			case "PUB_SUB_PUBLISHER_URL":
				errs += fmt.Sprintf("  Invalid PUB_SUB_PUBLISHER_URL: '%v'is not a valid URL\n", fe.Value())
			case "ELASTICSEARCH_URL":
				errs += fmt.Sprintf("  Invalid ELASTICSEARCH_URL: '%v' is not a valid URL\n", fe.Value())
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
		return nil, fmt.Errorf("Unable to parse BASE_URL '%s': %v", config.BaseURLString, err)
	}
	return &config, nil
}
