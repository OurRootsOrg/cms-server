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
	"reflect"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/codingconcepts/env"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/handlers"
	"github.com/hashicorp/logutils"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/server/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	"gocloud.dev/postgres"
)

const (
	defaultURL = "http://localhost:3001"
)

func main() {
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
	log.Printf("[INFO] env.ElasticsearchURLString: %s", env.ElasticsearchURLString)

	ap, err := api.NewAPI()
	if err != nil {
		log.Fatalf("Error calling NewAPI: %v", err)
	}
	defer ap.Close()
	ap = ap.ElasticsearchConfig(env.ElasticsearchURLString)

	// postgres
	log.Printf("Connecting to %s\n", env.DatabaseURL)
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
		CollectionPersister(p).
		RecordPersister(p)
	log.Print("[INFO] Using PostgresPersister")

	app := NewApp().BaseURL(*env.BaseURL).API(ap)
	if env.BaseURL.Scheme == "https" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

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
	BaseURL                *url.URL
	DatabaseURL            string `env:"DATABASE_URL" validate:"required,url"`
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
			case "DATABASE_URL":
				errs += fmt.Sprintf("  Invalid DATABASE_URL: '%v' is not a valid PostgreSQL URL\n", fe.Value())
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
		return nil, fmt.Errorf("unable to parse BASE_URL '%s': %v", config.BaseURLString, err)
	}
	return &config, nil
}
