package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ourrootsorg/cms-server/api"

	"github.com/gorilla/mux"
	"github.com/ourrootsorg/cms-server/model"
)

const contentType = "application/json"

// App is the container for the application
type App struct {
	baseURL url.URL
	api     api.LocalAPI
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
func (app *App) API(api api.LocalAPI) *App {
	app.api = api
	return app
}

// GetIndex redirects to the Swagger documentation
func (app App) GetIndex(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req,
		app.baseURL.Path+"/swagger/index.html?oauth2RedirectUrl="+url.QueryEscape(app.baseURL.String()+"/swagger/oauth2-redirect.html")+
			"&url="+url.QueryEscape(app.baseURL.String()+"/swagger/doc.json"), http.StatusTemporaryRedirect)
}

// GetHealth always returns `http.StatusOK` to indicate a running server
func (app App) GetHealth(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// NewRouter builds a router for handling requests
func (app App) NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc(app.baseURL.Path+"/", app.GetIndex).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/health", app.GetHealth).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/index.html", app.GetIndex).Methods("GET")

	r.HandleFunc(app.baseURL.Path+"/search", app.Search).Methods("GET")

	return r
}

// NotFound returns an http.StatusNotFound response
func NotFound(w http.ResponseWriter, req *http.Request) {
	m := fmt.Sprintf("Path '%s' not found", req.URL.RequestURI())
	log.Print("[ERROR] " + m)
	ErrorResponse(w, http.StatusNotFound, m)
}

func serverError(w http.ResponseWriter, err error) {
	log.Print("[ERROR] Server error: " + err.Error())
	// debug.PrintStack()
	ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err.Error()))
}

// ErrorResponse returns an error response
func ErrorResponse(w http.ResponseWriter, code int, message string) {
	ErrorsResponse(w, model.NewErrors(code, model.NewError(model.ErrOther, message)))
}

// ErrorsResponse returns an HTTP response from a model.Errors
func ErrorsResponse(w http.ResponseWriter, errors *model.Errors) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(errors.HTTPStatus())
	enc := json.NewEncoder(w)
	err := enc.Encode(errors.Errs())
	if err != nil {
		log.Printf("[ERROR] Failure encoding error response: '%v'", err)
	}
}
