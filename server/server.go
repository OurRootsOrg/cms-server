//go:generate sh -c "swag init --dir $(greadlink -f $(dirname $GOFILE)/../) --generalInfo server/$(basename $GOFILE)"
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jancona/ourroots/model"
	"github.com/jancona/ourroots/persist"
	"github.com/jancona/ourroots/server/docs"
	"gocloud.dev/postgres"

	// _ "github.com/jackc/pgx/v4/stdlib"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	contentType = "application/json"
	defaultURL  = "http://localhost:3000"
	indexHTML   = `<html>
	<body>
		<a href='/swagger/'>Swagger API Documentation</a><br/>
	</body>
</html>
`
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
	// Find out if we're running in a Lambda function
	isLambda := true
	if os.Getenv("LAMBDA_TASK_ROOT") == "" {
		isLambda = false
	}
	// baseURL is used to build proper absolute URLs in a couple of places
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = defaultURL
	}
	log.Printf("BaseURL: %s", baseURL)
	reqURL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		log.Fatalf("Error parsing base URL '%s': %v", baseURL, err)
	}
	app := App{
		BaseURL: *reqURL,
	}
	model.Initialize(app.BaseURL.Path)
	switch os.Getenv("PERSISTER") {
	case "sql":
		// db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
		db, err := postgres.Open(context.TODO(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
				err,
				os.Getenv("DATABASE_URL"),
			)
		}
		p := persist.NewPostgresPersister(app.BaseURL.Path, db)
		app.CategoryPersister = p
		app.CollectionPersister = p
		log.Print("Using PostgresPersister")
	case "memory":
		p := persist.NewMemoryPersister(app.BaseURL.Path)
		app.CategoryPersister = p
		app.CollectionPersister = p
		log.Print("Using MemoryPersister")
	default:
		log.Fatalf("Invalid PERSISTER: '%s'. Valid choices are 'sql' or 'memory'.", os.Getenv("PERSISTER"))
	}
	r := NewRouter(app)
	docs.SwaggerInfo.Host = reqURL.Hostname()
	if reqURL.Port() != "" {
		docs.SwaggerInfo.Host += ":" + reqURL.Port()
	}
	docs.SwaggerInfo.BasePath = reqURL.Path
	r.PathPrefix(reqURL.Path + "/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(baseURL+"/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))
	r.NotFoundHandler = http.HandlerFunc(notFound)

	if isLambda {
		// Lambda-specific setup
		// Note that the Lamda doesn't serve static content, only the API
		// API Gateway proxies static content requests directly to an S3 bucket
		// API Gateway + Lambda is https-only
		docs.SwaggerInfo.Schemes = []string{"https"}
		adapter := gorillamux.New(r)
		lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			log.Printf("Lambda request %#v", req)
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
		log.Fatal(http.ListenAndServe(reqURL.Host,
			handlers.LoggingHandler(os.Stdout,
				handlers.CORS(
					handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
					handlers.AllowedMethods([]string{"GET", "POST", "PATCH", "DELETE", "HEAD", "OPTIONS"}),
					handlers.AllowedOrigins([]string{"*"}))(r))))
	}
}

// NewRouter builds a router for handling requests
func NewRouter(app App) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc(app.BaseURL.Path+"/", app.GetIndex).Methods("GET")
	r.HandleFunc(app.BaseURL.Path+"/index.html", app.GetIndex).Methods("GET")

	r.HandleFunc(app.BaseURL.Path+"/categories", app.GetAllCategories).Methods("GET")
	r.HandleFunc(app.BaseURL.Path+"/categories", app.PostCategory).Methods("POST")
	r.HandleFunc(app.BaseURL.Path+"/categories/{id}", app.GetCategory).Methods("GET")
	r.HandleFunc(app.BaseURL.Path+"/categories/{id}", app.PatchCategory).Methods("PATCH")
	r.HandleFunc(app.BaseURL.Path+"/categories/{id}", app.DeleteCategory).Methods("DELETE")

	r.HandleFunc(app.BaseURL.Path+"/collections", app.GetAllCollections).Methods("GET")
	r.HandleFunc(app.BaseURL.Path+"/collections", app.PostCollection).Methods("POST")
	r.HandleFunc(app.BaseURL.Path+"/collections/{id}", app.GetCollection).Methods("GET")
	r.HandleFunc(app.BaseURL.Path+"/collections/{id}", app.PatchCollection).Methods("PATCH")
	r.HandleFunc(app.BaseURL.Path+"/collections/{id}", app.DeleteCollection).Methods("DELETE")
	return r
}

// App is the container for the application
type App struct {
	CategoryPersister   model.CategoryPersister
	CollectionPersister model.CollectionPersister
	BaseURL             url.URL
	// PathPrefix          string
}

// GetIndex returns an HTML index page
func (app App) GetIndex(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}

func serverError(w http.ResponseWriter, err error) {
	log.Print("Server error: " + err.Error())
	// debug.PrintStack()
	errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err.Error()))
}

func notFound(w http.ResponseWriter, req *http.Request) {
	m := fmt.Sprintf("Path '%s' not found", req.URL.RequestURI())
	log.Print(m)
	errorResponse(w, http.StatusNotFound, m)
}

func errorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	e := model.Error{Code: code, Message: message}
	err := enc.Encode(e)
	if err != nil {
		log.Printf("Failure encoding error response: '%v'", err)
	}
}
