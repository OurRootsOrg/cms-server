//go:generate sh -c "swag init --dir $(dirname $(pwd)) --output ../api/docs --generalInfo server/$GOFILE"
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/handlers"
	"github.com/hashicorp/logutils"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/api/docs"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"gocloud.dev/postgres"

	// _ "github.com/jackc/pgx/v4/stdlib"
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
func main() {
	env, err := ParseEnv()
	if err != nil {
		log.Fatalf("Error parsing environmet variables: %v", err)
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel(env.MinLogLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	log.Printf("[INFO] env.BaseURLString: %s, env.BaseURL.Path: %s", env.BaseURLString, env.BaseURL.Path)
	model.Initialize(env.BaseURL.Path)
	app := api.NewApp().BaseURL(*env.BaseURL)
	switch env.Persister {
	case "sql":
		db, err := postgres.Open(context.TODO(), env.DatabaseURL)
		if err != nil {
			log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
				err,
				env.DatabaseURL,
			)
		}
		p := persist.NewPostgresPersister(env.BaseURL.Path, db)
		app.CategoryPersister(p).CollectionPersister(p)
		log.Print("[INFO] Using PostgresPersister")
	case "memory":
		p := persist.NewMemoryPersister(env.BaseURL.Path)
		app.CategoryPersister(p).CollectionPersister(p)
		log.Print("[INFO] Using MemoryPersister")
	default:
		// Should never happen
		log.Fatalf("Invalid PERSISTER: '%s', valid choices are 'sql' or 'memory'.", env.Persister)
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
	r.NotFoundHandler = http.HandlerFunc(api.NotFound)

	if env.IsLambda {
		// Lambda-specific setup
		// Note that the Lamda doesn't serve static content, only the API
		// API Gateway proxies static content requests directly to an S3 bucket
		// API Gateway + Lambda is https-only
		docs.SwaggerInfo.Schemes = []string{"https"}
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
		log.Fatal(http.ListenAndServe(env.BaseURL.Host,
			handlers.LoggingHandler(os.Stdout,
				handlers.CORS(
					handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
					handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}),
					handlers.AllowedOrigins([]string{"*"}))(r))))
	}
}

// Env holds values parse from environment variables
type Env struct {
	IsLambda      bool   `env:"LAMBDA_TASK_ROOT"`
	MinLogLevel   string `env:"MIN_LOG_LEVEL" validate:"omitempty,eq=DEBUG|eq=INFO|eq=ERROR"`
	BaseURLString string `env:"BASE_URL" validate:"omitempty,url"`
	Persister     string `env:"PERSISTER" validate:"required,eq=memory|eq=sql"`
	DatabaseURL   string `env:"DATABASE_URL" validate:"omitempty,url"`
	BaseURL       *url.URL
}

// ParseEnv parses and validates environment variables and stores them in the Env structure
func ParseEnv() (*Env, error) {
	env := Env{
		IsLambda:      os.Getenv("LAMBDA_TASK_ROOT") != "",
		MinLogLevel:   os.Getenv("MIN_LOG_LEVEL"),
		BaseURLString: os.Getenv("BASE_URL"),
		Persister:     os.Getenv("PERSISTER"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
	}
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("env")
	})
	err := validate.Struct(env)
	if err != nil {
		errs := "Error parsing environment variables:\n"
		for _, fe := range err.(validator.ValidationErrors) {
			switch fe.Field() {
			case "MIN_LOG_LEVEL":
				errs += fmt.Sprintf("  Invalid MIN_LOG_LEVEL: '%v', valid values are 'DEBUG', 'INFO' or 'ERROR'\n", fe.Value())
			case "BASE_URL":
				errs += fmt.Sprintf("  Invalid BASE_URL: '%v' is not a valid URL\n", fe.Value())
			case "PERSISTER":
				if fe.Tag() == "required" {
					errs += "  PERSISTER is required, valid values are 'memory' or 'sql'\n"
				} else {
					errs += fmt.Sprintf("  Invalid PERSISTER: '%v', valid values are 'memory' or 'sql'\n", fe.Value())
				}
			case "DATABASE_URL":
				errs += fmt.Sprintf("  Invalid DATABASE_URL: '%v' is not a valid Postgresql URL\n", fe.Value())
			}
		}
		return nil, errors.New(errs)
	}
	if env.MinLogLevel == "" {
		env.MinLogLevel = "DEBUG"
	}
	if env.BaseURLString == "" {
		env.BaseURLString = defaultURL
	}
	env.BaseURL, err = url.ParseRequestURI(env.BaseURLString)
	if err != nil {
		// Unreachable, if the validator does its job
		return nil, fmt.Errorf("Unable to parse BASE_URL '%s': %v", env.BaseURLString, err)
	}
	if env.Persister == "sql" && env.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required for PERSISTER=sql")
	}
	return &env, nil
}
