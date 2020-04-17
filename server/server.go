//go:generate swag init -g $GOFILE
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jancona/ourroots/model"
	"github.com/jancona/ourroots/server/docs"

	// "github.com/jancona/ourroots/server/docs"
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
	url, err := url.ParseRequestURI(baseURL)
	if err != nil {
		log.Fatalf("Error parsing base URL '%s': %v", baseURL, err)
	}
	log.Printf("BaseURL: %s, url: %s", baseURL, url)
	r := NewRouter(App{
		Categories: make(map[uuid.UUID]model.Category),
		BaseURL:    baseURL,
	})
	docs.SwaggerInfo.Host = url.Hostname()
	if url.Port() != "" {
		docs.SwaggerInfo.Host += ":" + url.Port()
	}
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
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
		log.Fatal(http.ListenAndServe(url.Host,
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
	r.HandleFunc("/", app.GetIndex).Methods("GET")
	r.HandleFunc("/index.html", app.GetIndex).Methods("GET")
	r.HandleFunc("/categories", app.GetAllCategories).Methods("GET")
	r.HandleFunc("/categories", app.PostCategory).Methods("POST")
	r.HandleFunc("/categories", app.DeleteAllCategories).Methods("DELETE")
	r.HandleFunc("/categories/{id}", app.GetCategory).Methods("GET")
	r.HandleFunc("/categories/{id}", app.PatchCategory).Methods("PATCH")
	r.HandleFunc("/categories/{id}", app.DeleteCategory).Methods("DELETE")
	return r
}

// App is the container for the application
type App struct {
	BaseURL string
	// Dummy "database".
	// (Note that this behaves really strangely when multiple Lambda are running
	// because which database you see depends on which Lamda instance you're routed to.)
	Categories map[uuid.UUID]model.Category
}

// GetIndex returns an HTML index page
func (app App) GetIndex(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}

// GetAllCategories returns all categories in the database
// @summary returns all categories
// @router /categories [get]
// @tags categories
// @id find
// @Param dummy body model.FieldDefSet false "Dummy"
// @Param dummy body model.fieldType false "Dummy"
// @produce application/json
// @success 200 {array} model.Category "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) GetAllCategories(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	v := make([]model.Category, 0, len(app.Categories))

	for _, value := range app.Categories {
		v = append(v, value)
	}
	err := enc.Encode(v)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeleteAllCategories deletes all Categories from the database
// @summary deletes all Categories
// @router /categories [delete]
// @tags categories
// @id deleteAll
// @success 200 {array} model.Category "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) DeleteAllCategories(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)

	for id := range app.Categories {
		delete(app.Categories, id)
	}
	v := make([]model.Category, 0)
	err := enc.Encode(v)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PostCategory adds a new Category to the database
// @summary adds a new Category
// @router /categories [post]
// @tags categories
// @id addOne
// @Param category body model.CategoryInput true "Add Category"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Category "OK"
// @failure 415 {object} model.Error "Bad Content-Type"
// @failure 500 {object} model.Error "Server error"
func (app App) PostCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	if mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	ci, _ := model.NewCategoryInput("")
	err = json.NewDecoder(req.Body).Decode(&ci)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	category := model.NewCategory(ci)
	// Add to "database"
	app.Categories[category.ID] = category
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetCategory gets a Category from the database
// @summary gets a Category
// @router /categories/{id} [get]
// @tags categories
// @id getOne
// @Param id path string true "Category ID" format(uuid)
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 404 {object} model.Error "Not found"
// @failure 500 {object} model.Error "Server error"
func (app App) GetCategory(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	vars := mux.Vars(req)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		notFound(w, req)
		return
	}
	category, found := app.Categories[id]
	if !found {
		notFound(w, req)
		return
	}
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PatchCategory updates a Category in the database
// @summary updates a Category
// @router /categories/{id} [patch]
// @tags categories
// @id update
// @Param id path string true "Category ID" format(uuid)
// @Param category body model.CategoryInput true "Update Category"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 415 {object} model.Error "Bad Content-Type"
// @failure 500 {object} model.Error "Server error"
func (app App) PatchCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		serverError(w, err)
		return
	}
	vars := mux.Vars(req)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		notFound(w, req)
		return
	}
	_, found := app.Categories[id]
	if !found {
		// Not allowed to add a Category with PATCH
		notFound(w, req)
		return
	}
	if mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	var tdi model.CategoryInput
	err = json.NewDecoder(req.Body).Decode(&tdi)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	category := app.Categories[id]
	category.Name = tdi.Name
	// Add to "database"
	app.Categories[id] = category
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeleteCategory deletes a Category from the database
// @summary deletes a Category
// @router /categories/{id} [delete]
// @tags categories
// @id delete
// @Param id path string true "Category ID" format(uuid)
// @success 204 {object} model.Category "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) DeleteCategory(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		notFound(w, req)
		return
	}
	delete(app.Categories, id)
	w.WriteHeader(http.StatusNoContent)
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
