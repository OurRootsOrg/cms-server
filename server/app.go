package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/ourrootsorg/cms-server/api"
)

const contentType = "application/json"

// App is the container for the application
type App struct {
	baseURL url.URL
	api     *api.API
}

// NewApp builds an App
func NewApp() *App {
	app := &App{
		baseURL: url.URL{},
	}
	return app
}

// BaseURL sets the base URL for the app
func (app *App) BaseURL(url url.URL) *App {
	app.baseURL = url
	return app
}

// API sets the API object for the app
func (app *App) API(api *api.API) *App {
	log.Printf("[DEBUG] api: %#v", app.api)
	app.api = api
	return app
}

// GetIndex redirects to the Swagger documentation
func (app App) GetIndex(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, app.baseURL.Path+"/swagger/", http.StatusTemporaryRedirect)
}

// NewRouter builds a router for handling requests
func (app App) NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc(app.baseURL.Path+"/", app.GetIndex).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/index.html", app.GetIndex).Methods("GET")

	r.HandleFunc(app.baseURL.Path+"/categories", app.GetAllCategories).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/categories", app.PostCategory).Methods("POST")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.GetCategory).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.PutCategory).Methods("PUT")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.DeleteCategory).Methods("DELETE")

	r.HandleFunc(app.baseURL.Path+"/collections", app.GetCollections).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/collections", app.PostCollection).Methods("POST")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.GetCollection).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.PutCollection).Methods("PUT")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.DeleteCollection).Methods("DELETE")
	return r
}

// NotFound returns an http.StatusNotFound response
func NotFound(w http.ResponseWriter, req *http.Request) {
	m := fmt.Sprintf("Path '%s' not found", req.URL.RequestURI())
	log.Print("[ERROR] " + m)
	OtherErrorResponse(w, http.StatusNotFound, m)
}

func serverError(w http.ResponseWriter, err error) {
	log.Print("[ERROR] Server error: " + err.Error())
	// debug.PrintStack()
	OtherErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err.Error()))
}

// OtherErrorResponse returns an error response
func OtherErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	e := api.NewError(api.ErrOther, message)
	err := enc.Encode([]api.Error{e})
	if err != nil {
		log.Printf("[ERROR] Failure encoding error response: '%v'", err)
	}
}

// ValidationErrorResponse returns a validation error response
func ValidationErrorResponse(w http.ResponseWriter, code int, er error) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	errors := api.NewErrors(http.StatusBadRequest, er)
	log.Printf("[DEBUG] errBody: %#v", errors)
	err := enc.Encode(errors.Errs())
	if err != nil {
		log.Printf("[ERROR] Failure encoding error response: '%v'", err)
	}
}

// ErrorsResponse returns an HTTP response from a api.Errors
func ErrorsResponse(w http.ResponseWriter, errors *api.Errors) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(errors.HTTPStatus())
	enc := json.NewEncoder(w)
	err := enc.Encode(errors.Errs())
	if err != nil {
		log.Printf("[ERROR] Failure encoding error response: '%v'", err)
	}
}
