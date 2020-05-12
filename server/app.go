package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
)

const contentType = "application/json"

type localAPI interface {
	GetCategories(context.Context) (*api.CategoryResult, *model.Errors)
	GetCategory(ctx context.Context, id string) (*model.Category, *model.Errors)
	AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, *model.Errors)
	UpdateCategory(ctx context.Context, id string, in model.Category) (*model.Category, *model.Errors)
	DeleteCategory(ctx context.Context, id string) *model.Errors
	GetCollections(ctx context.Context /* filter/search criteria */) (*api.CollectionResult, *model.Errors)
	GetCollection(ctx context.Context, id string) (*model.Collection, *model.Errors)
	AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, *model.Errors)
	UpdateCollection(ctx context.Context, id string, in model.Collection) (*model.Collection, *model.Errors)
	DeleteCollection(ctx context.Context, id string) *model.Errors
	GetPosts(ctx context.Context /* filter/search criteria */) (*api.PostResult, *model.Errors)
	GetPost(ctx context.Context, id string) (*model.Post, *model.Errors)
	AddPost(ctx context.Context, in model.PostIn) (*model.Post, *model.Errors)
	UpdatePost(ctx context.Context, id string, in model.Post) (*model.Post, *model.Errors)
	DeletePost(ctx context.Context, id string) *model.Errors
	PostContentRequest(ctx context.Context, contentRequest api.ContentRequest) (*api.ContentResult, *model.Errors)
	GetContent(ctx context.Context, key string) ([]byte, *model.Errors)
}

// App is the container for the application
type App struct {
	baseURL      url.URL
	api          localAPI
	oidcAudience string
	oidcDomain   string
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
func (app *App) API(api localAPI) *App {
	app.api = api
	return app
}

// OIDCAudience sets the OIDC audience for the app
func (app *App) OIDCAudience(oidcAudience string) *App {
	app.oidcAudience = oidcAudience
	return app
}

// OIDCDomain sets the OIDC domain for the app
func (app *App) OIDCDomain(oidcDomain string) *App {
	app.oidcDomain = oidcDomain
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
	// if app.oidcAudience == "" || app.oidcDomain == "" {
	// 	// Run without authentication
	// }
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			log.Printf("[DEBUG] token: %#v, oidcAudience: %#v, oidcDomain: %#v", *token, app.oidcAudience, app.oidcDomain)
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(app.oidcAudience, false)
			if !checkAud {
				log.Printf("[DEBUG] invalid audience %v", *token)
				return token, errors.New("invalid audience")
			}
			// Verify 'iss' claim
			iss := "https://" + app.oidcDomain + "/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				log.Printf("[DEBUG] invalid issuer %v", *token)
				return token, errors.New("invalid issuer")
			}

			cert, err := getPemCert(token)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc(app.baseURL.Path+"/", app.GetIndex).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/health", app.GetHealth).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/index.html", app.GetIndex).Methods("GET")

	r.Handle(app.baseURL.Path+"/categories", jwtMiddleware.Handler(http.HandlerFunc(app.GetAllCategories))).Methods("GET")
	// r.HandleFunc(app.baseURL.Path+"/categories", app.GetAllCategories).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/categories", app.PostCategory).Methods("POST")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.GetCategory).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.PutCategory).Methods("PUT")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.DeleteCategory).Methods("DELETE")

	r.HandleFunc(app.baseURL.Path+"/collections", app.GetCollections).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/collections", app.PostCollection).Methods("POST")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.GetCollection).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.PutCollection).Methods("PUT")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.DeleteCollection).Methods("DELETE")

	r.HandleFunc(app.baseURL.Path+"/content", app.PostContentRequest).Methods("POST")

	r.HandleFunc(app.baseURL.Path+"/posts", app.GetPosts).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/posts", app.PostPost).Methods("POST")
	r.HandleFunc(app.baseURL.Path+"/posts/{id}", app.GetPost).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/posts/{id}", app.PutPost).Methods("PUT")
	r.HandleFunc(app.baseURL.Path+"/posts/{id}", app.DeletePost).Methods("DELETE")

	return r
}

type jwks struct {
	Keys []jsonWebKeys `json:"keys"`
}

type jsonWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get("https://" + os.Getenv("AUTH0_DOMAIN") + "/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwk = jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwk)

	if err != nil {
		return cert, err
	}

	for k := range jwk.Keys {
		if token.Header["kid"] == jwk.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwk.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
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
