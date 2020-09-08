package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/go-oidc"
)

const contentType = "application/json"

// verifier allows use of a mock verifier for testing
type verifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

// App is the container for the application
type App struct {
	baseURL      url.URL
	api          api.LocalAPI
	oidcAudience string
	oidcDomain   string
	oidcProvider *oidc.Provider
	oidcVerifier verifier
	authDisabled bool // If set to true, this disables authentication. This should only be done in test code!
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

// OIDC sets up OIDC configuration for the app
func (app *App) OIDC(oidcAudience string, oidcDomain string) *App {
	var err error
	app.oidcAudience = oidcAudience
	app.oidcDomain = oidcDomain
	// Assumes that the OIDC provider supports discovery
	app.oidcProvider, err = oidc.NewProvider(context.TODO(), app.oidcDomain)
	if err != nil {
		log.Fatalf("Unable to intialize OIDC verifier: %v", err)
	}
	config := &oidc.Config{
		ClientID: app.oidcAudience,
	}
	app.oidcVerifier = app.oidcProvider.Verifier(config)
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

// we need a handler for options. This handler won't actually get invoked; it's just needed so the CORS middleware will get invoked
func (app App) OptionsNoop(w http.ResponseWriter, req *http.Request) {
}

func (app App) verifyToken(next http.Handler) http.Handler {
	if app.authDisabled {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	fn := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			msg := "No Authorization header found"
			log.Print("[DEBUG] " + msg)
			ErrorResponse(w, http.StatusUnauthorized, msg)
			return
		}
		authHeaderParts := strings.Fields(authHeader)
		if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
			msg := "Authorization header format must be Bearer {token}"
			log.Print("[DEBUG] " + msg)
			ErrorResponse(w, http.StatusUnauthorized, msg)
			return
		}

		// Make sure that the incoming request has our token header
		accessJWT := authHeaderParts[1]

		// Verify the access token
		ctx := r.Context()
		parsedToken, err := app.oidcVerifier.Verify(ctx, accessJWT)
		if err != nil {
			msg := fmt.Sprintf("Invalid token: %s", err.Error())
			log.Print("[DEBUG] " + msg)
			ErrorResponse(w, http.StatusUnauthorized, msg)
			return
		}
		log.Printf("[DEBUG] Found valid token for subject '%s'", parsedToken.Subject)
		user, errors := app.api.RetrieveUser(r.Context(), app.oidcProvider, parsedToken, accessJWT)
		if errors != nil {
			msg := fmt.Sprintf("RetrieveUser error %v", errors)
			log.Print("[ERROR] " + msg)
			ErrorsResponse(w, errors)
			return
		}

		// If we get here, everything worked and we can set the
		// user property in context.
		// c := context.WithValue(r.Context(), api.TokenProperty, parsedToken)
		c := context.WithValue(r.Context(), api.UserProperty, user)

		newRequest := r.WithContext(c)
		// Update the current request with the new context information.
		*r = *newRequest
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// NewRouter builds a router for handling requests
func (app App) NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc(app.baseURL.Path+"/", app.GetIndex).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/health", app.GetHealth).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/index.html", app.GetIndex).Methods("GET")

	r.Handle(app.baseURL.Path+"/categories", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/categories", app.verifyToken(http.HandlerFunc(app.GetAllCategories))).Methods("GET")
	r.Handle(app.baseURL.Path+"/categories", app.verifyToken(http.HandlerFunc(app.PostCategory))).Methods("POST")

	r.Handle(app.baseURL.Path+"/categories/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/categories/{id}", app.verifyToken(http.HandlerFunc(app.GetCategory))).Methods("GET")
	r.Handle(app.baseURL.Path+"/categories/{id}", app.verifyToken(http.HandlerFunc(app.PutCategory))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/categories/{id}", app.verifyToken(http.HandlerFunc(app.DeleteCategory))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/collections", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/collections", app.verifyToken(http.HandlerFunc(app.GetCollections))).Methods("GET")
	r.Handle(app.baseURL.Path+"/collections", app.verifyToken(http.HandlerFunc(app.PostCollection))).Methods("POST")

	r.Handle(app.baseURL.Path+"/collections/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/collections/{id}", app.verifyToken(http.HandlerFunc(app.GetCollection))).Methods("GET")
	r.Handle(app.baseURL.Path+"/collections/{id}", app.verifyToken(http.HandlerFunc(app.PutCollection))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/collections/{id}", app.verifyToken(http.HandlerFunc(app.DeleteCollection))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/content", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/content", app.verifyToken(http.HandlerFunc(app.PostContentRequest))).Methods("POST")

	r.Handle(app.baseURL.Path+"/posts", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/posts", app.verifyToken(http.HandlerFunc(app.GetPosts))).Methods("GET")
	r.Handle(app.baseURL.Path+"/posts", app.verifyToken(http.HandlerFunc(app.PostPost))).Methods("POST")

	r.Handle(app.baseURL.Path+"/posts/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/posts/{id}", app.verifyToken(http.HandlerFunc(app.GetPost))).Methods("GET")
	r.Handle(app.baseURL.Path+"/posts/{id}", app.verifyToken(http.HandlerFunc(app.PutPost))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/posts/{id}", app.verifyToken(http.HandlerFunc(app.DeletePost))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/posts/{id}/images/{filePath:.*}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/posts/{id}/images/{filePath:.*}", app.verifyToken(http.HandlerFunc(app.GetPostImage))).Methods("GET")

	r.Handle(app.baseURL.Path+"/records", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/records", app.verifyToken(http.HandlerFunc(app.GetRecords))).Methods("GET")

	r.Handle(app.baseURL.Path+"/records/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/records/{id}", app.verifyToken(http.HandlerFunc(app.GetRecord))).Methods("GET")

	r.Handle(app.baseURL.Path+"/settings", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/settings", app.verifyToken(http.HandlerFunc(app.GetSettings))).Methods("GET")
	r.Handle(app.baseURL.Path+"/settings", app.verifyToken(http.HandlerFunc(app.PutSettings))).Methods("PUT")

	// search doesn't require a token for now
	r.Handle(app.baseURL.Path+"/search", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.HandleFunc(app.baseURL.Path+"/search", app.Search).Methods("GET")

	r.Handle(app.baseURL.Path+"/search/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.HandleFunc(app.baseURL.Path+"/search/{id}", app.SearchByID).Methods("GET")

	r.Handle(app.baseURL.Path+"/places", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.HandleFunc(app.baseURL.Path+"/places", http.HandlerFunc(app.GetPlacesByPrefix)).Methods("GET")

	r.Handle(app.baseURL.Path+"/currentuser", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/currentuser", app.verifyToken(http.HandlerFunc(app.GetCurrentUser))).Methods("GET")

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
	ErrorsResponse(w, api.NewHTTPError(errors.New(message), code))
}

// ErrorsResponse returns an HTTP response from a api.Error
func ErrorsResponse(w http.ResponseWriter, err error) {
	var errors *api.Error
	var ok bool
	if errors, ok = err.(*api.Error); !ok {
		log.Printf("[INFO] Unexpectedly received an `error` instead of a `*api.Error`: '%v'", err)
		errors = api.NewError(err)
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(errors.HTTPStatus())
	enc := json.NewEncoder(w)
	err = enc.Encode(errors.Errs())
	if err != nil {
		log.Printf("[ERROR] Failure encoding error response: '%v'", err)
	}
}

// get a "id" variable from the request and validate > 0
func getIDFromRequest(req *http.Request) (uint32, error) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		return 0, api.NewError(fmt.Errorf("Bad id '%s': %v", vars["id"], err))
	}
	return uint32(id), nil
}
