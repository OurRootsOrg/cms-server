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
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/handlers"
	"github.com/hashicorp/logutils"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/server/docs"
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
	ap := api.NewAPI().BaseURL(*env.BaseURL)
	app := NewApp().BaseURL(*env.BaseURL).API(ap)

	switch env.Persister {
	case "sql":
		log.Printf("Connecting to %s\n", env.DatabaseURL)
		db, err := postgres.Open(context.TODO(), env.DatabaseURL)
		if err != nil {
			log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
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
			log.Fatalf("Error connecting to database: %v\n DATABASE_URL: %s\n",
				err,
				env.DatabaseURL,
			)
		}
		log.Printf("Connected to %s\n", env.DatabaseURL)

		p := persist.NewPostgresPersister(env.BaseURL.Path, db)
		ap.CategoryPersister(p).CollectionPersister(p)
		log.Print("[INFO] Using PostgresPersister")
	case "memory":
		p := persist.NewMemoryPersister(env.BaseURL.Path)
		ap.CategoryPersister(p).CollectionPersister(p)
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
	r.NotFoundHandler = http.HandlerFunc(NotFound)

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
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.BaseURL.Port()), // changed from Host because host doesn't work inside docker
			handlers.LoggingHandler(os.Stdout,
				handlers.CORS(
					handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
					handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}),
					handlers.AllowedOrigins([]string{"*"}))(r))))
	}
}

// Env holds values parse from environment variables
type Env struct {
	LambdaTaskRoot string   `envconfig:"LAMBDA_TASK_ROOT"`
	IsLambda       bool     `ignored:"true"`
	MinLogLevel    string   `envconfig:"MIN_LOG_LEVEL"`
	Persister      string   `envconfig:"PERSISTER" required:"true"`
	DatabaseURL    string   `envconfig:"DATABASE_URL"`
	BaseURLString  string   `envconfig:"BASE_URL"`
	BaseURL        *url.URL `ignored:"true"`
}

// ParseEnv parses and validates environment variables and stores them in the Env structure
func ParseEnv() (*Env, error) {
	var env Env
	err := envconfig.Process("", &env)
	if err != nil {
		return nil, err
	}
	env.IsLambda = env.LambdaTaskRoot != ""
	errs := ""
	if env.MinLogLevel == "" {
		env.MinLogLevel = "DEBUG"
	}
	if env.MinLogLevel != "DEBUG" && env.MinLogLevel != "INFO" && env.MinLogLevel != "ERROR" {
		errs += fmt.Sprintf("  Invalid MIN_LOG_LEVEL: '%v', valid values are 'DEBUG', 'INFO' or 'ERROR'\n", env.MinLogLevel)
	}
	if env.Persister != "memory" && env.Persister != "sql" {
		errs += fmt.Sprintf("  Invalid PERSISTER: '%v', valid values are 'memory' or 'sql'\n", env.Persister)
	}
	if env.Persister == "sql" && env.DatabaseURL == "" {
		errs += "DATABASE_URL is required for PERSISTER=sql"
	}
	if _, err := url.ParseRequestURI(env.DatabaseURL); env.DatabaseURL != "" && err != nil {
		errs += fmt.Sprintf("Unable to parse BASE_URL '%s': %v", env.DatabaseURL, err)
	}
	if env.BaseURLString == "" {
		env.BaseURLString = defaultURL
	}
	env.BaseURL, err = url.ParseRequestURI(env.BaseURLString)
	if err != nil {
		errs += fmt.Sprintf("Unable to parse BASE_URL '%s': %v", env.BaseURLString, err)
	}
	if errs != "" {
		return nil, errors.New(errs)
	}
	return &env, nil
}
