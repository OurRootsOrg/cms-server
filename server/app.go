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
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/ourrootsorg/cms-server/model"

	"github.com/ourrootsorg/cms-server/utils"

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
	baseURL          url.URL
	api              api.LocalAPI
	oidcAudience     string
	oidcDomain       string
	oidcProvider     *oidc.Provider
	oidcVerifier     verifier
	sandboxSocietyID uint32
	authDisabled     bool // If set to true, this disables authentication. This should only be done in test code!
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

// SandboxSociety - everyone is added as an editor to the default society if it exists
func (app *App) SandboxSociety(id uint32) *App {
	app.sandboxSocietyID = id
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

func (app App) setSociety(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		societyID, err := getSocietyIDFromRequest(r)
		if err != nil || societyID <= 0 {
			ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Bad society id %v", err))
			return
		}

		c := utils.AddSocietyIDToContext(r.Context(), uint32(societyID))
		newRequest := r.WithContext(c)
		// Update the current request with the new context information.
		*r = *newRequest
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
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
		//log.Printf("[DEBUG] Found valid token for subject '%s'", parsedToken.Subject)
		user, isNew, errors := app.api.RetrieveUser(r.Context(), app.oidcProvider, parsedToken, accessJWT)
		if errors != nil {
			msg := fmt.Sprintf("RetrieveUser error %v", errors)
			log.Print("[ERROR] " + msg)
			ErrorsResponse(w, errors)
			return
		}

		// If we get here, everything worked and we can set the
		// user property in context.
		// c := context.WithValue(r.Context(), api.TokenProperty, parsedToken)
		c := utils.AddUserToContext(r.Context(), user)

		// if the user is new and there is a default society, add the user as an editor of the society
		if isNew && app.sandboxSocietyID > 0 {
			err = app.addUserToSandboxSociety(c, model.AuthEditor)
			if err != nil {
				msg := fmt.Sprintf("AddSocietyUser to sandbox society error %v", err)
				log.Print("[ERROR] " + msg)
				ErrorsResponse(w, err)
				return
			}
		}

		newRequest := r.WithContext(c)
		// Update the current request with the new context information.
		*r = *newRequest
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (app App) verifySearchToken(next http.Handler) http.Handler {
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
		token, err := jwt.Parse(accessJWT, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			societyID, _, err := parseSearchTokenClaims(token)
			if err != nil {
				return nil, err
			}
			society, err := app.api.GetSociety(ctx, societyID)
			if err != nil {
				return nil, err
			}
			return []byte(society.SecretKey), nil
		})
		errMsg := ""
		if err != nil || !token.Valid {
			errMsg = fmt.Sprintf("Invalid search token: %s %v", accessJWT, err)
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			errMsg = fmt.Sprintf("Search token claims of the wrong type: %s", accessJWT)
		}
		err = claims.Valid()
		if err != nil {
			errMsg = fmt.Sprintf("Invalid search token claims: %s %v", accessJWT, err)
		}
		if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
			errMsg = fmt.Sprintf("Search token is expired: %s", accessJWT)
		}
		if errMsg != "" {
			log.Print("[DEBUG] " + errMsg)
			ErrorResponse(w, http.StatusUnauthorized, errMsg)
			return
		}

		// save the society ID and user ID in context
		societyID, userID, err := parseSearchTokenClaims(token)
		ctx = utils.AddSocietyIDToContext(ctx, societyID)
		ctx = utils.AddSearchUserIDToContext(ctx, userID)

		newRequest := r.WithContext(ctx)
		// Update the current request with the new context information.
		*r = *newRequest
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func parseSearchTokenClaims(token *jwt.Token) (uint32, uint32, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, fmt.Errorf("invalid search claims")
	}
	subject, ok := claims["sub"].(string)
	if !ok {
		return 0, 0, fmt.Errorf("invalid search claims subject %v", subject)
	}
	subjectParts := strings.Split(subject, "_")
	if len(subjectParts) != 2 {
		return 0, 0, fmt.Errorf("unexpected subject: %s", subject)
	}
	societyID, err := strconv.Atoi(subjectParts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected subject: %s", subject)
	}
	userID, err := strconv.Atoi(subjectParts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected subject: %s", subject)
	}
	return uint32(societyID), uint32(userID), nil
}

func (app App) addUserToSandboxSociety(ctx context.Context, level model.AuthLevel) error {
	sctx := utils.AddSocietyIDToContext(ctx, app.sandboxSocietyID)
	_, err := app.api.GetSociety(sctx, app.sandboxSocietyID)
	if err != nil {
		msg := fmt.Sprintf("Get sandbox society error %v", err)
		log.Print("[ERROR] " + msg)
		// don't return an error, maybe sandbox society hasn't been created yet
		return nil
	}
	body := model.SocietyUserBody{
		Level: level,
	}
	_, err = app.api.AddSocietyUser(sctx, body)
	return err
}

func (app App) authenticate(minAuthLevel model.AuthLevel, next http.Handler) http.Handler {
	if app.authDisabled {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, err := utils.GetUserFromContext(ctx)
		if err != nil || user == nil {
			ErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("Missing user: %v", err))
			return
		}
		// read society user
		societyUser, errors := app.api.GetSocietyUserByUser(r.Context(), user.ID)
		if errors != nil {
			msg := fmt.Sprintf("RetrieveUserSociety error %v", errors)
			log.Print("[ERROR] " + msg)
			ErrorsResponse(w, errors)
			return
		}

		if societyUser.Level < minAuthLevel {
			ErrorResponse(w, http.StatusForbidden,
				fmt.Sprintf("User is level '%s' but '%s' is required: %v", societyUser.Level, minAuthLevel, err))
			return
		}
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

	r.Handle(app.baseURL.Path+"/societies/{society}/categories", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/categories", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetAllCategories))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}/categories", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.PostCategory))))).Methods("POST")

	r.Handle(app.baseURL.Path+"/societies/{society}/categories/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/categories/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetCategory))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}/categories/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.PutCategory))))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/societies/{society}/categories/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.DeleteCategory))))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/societies/{society}/collections", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/collections", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetCollections))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}/collections", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.PostCollection))))).Methods("POST")

	r.Handle(app.baseURL.Path+"/societies/{society}/collections/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/collections/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetCollection))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}/collections/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.PutCollection))))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/societies/{society}/collections/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.DeleteCollection))))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/societies/{society}/content", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/content", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.PostContentRequest))))).Methods("POST")

	r.Handle(app.baseURL.Path+"/societies/{society}/posts", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/posts", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetPosts))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}/posts", app.setSociety(app.verifyToken(app.authenticate(model.AuthContributor,
		http.HandlerFunc(app.PostPost))))).Methods("POST")

	r.Handle(app.baseURL.Path+"/societies/{society}/posts/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/posts/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetPost))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}/posts/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.PutPost))))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/societies/{society}/posts/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthEditor,
		http.HandlerFunc(app.DeletePost))))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/societies/{society}/posts/{id}/images/{filePath:.*}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/posts/{id}/images/{filePath:.*}", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetPostImage))))).Methods("GET")

	r.Handle(app.baseURL.Path+"/societies/{society}/records", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/records", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetRecords))))).Methods("GET")

	r.Handle(app.baseURL.Path+"/societies/{society}/records/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/records/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetRecord))))).Methods("GET")

	r.Handle(app.baseURL.Path+"/society_summaries", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/society_summaries", app.verifyToken(http.HandlerFunc(app.GetSocietySummariesForCurrentUser))).Methods("GET")

	r.Handle(app.baseURL.Path+"/society_summaries/{society}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/society_summaries/{society}", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetSocietySummary))))).Methods("GET")

	r.Handle(app.baseURL.Path+"/societies", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies", app.verifyToken(http.HandlerFunc(app.PostSociety))).Methods("POST")

	r.Handle(app.baseURL.Path+"/societies/{society}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.GetSociety))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.PutSociety))))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/societies/{society}", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.DeleteSociety))))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/societies/{society}/current_user", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/current_user", app.setSociety(app.verifyToken(app.authenticate(model.AuthReader,
		http.HandlerFunc(app.GetCurrentSocietyUser))))).Methods("GET")

	r.Handle(app.baseURL.Path+"/societies/{society}/users", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/users", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.GetSocietyUserNames))))).Methods("GET")

	r.Handle(app.baseURL.Path+"/societies/{society}/users/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/users/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.PutSocietyUserName))))).Methods("PUT")
	r.Handle(app.baseURL.Path+"/societies/{society}/users/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.DeleteSocietyUser))))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/societies/{society}/invitations", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/invitations", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.GetInvitations))))).Methods("GET")
	r.Handle(app.baseURL.Path+"/societies/{society}/invitations", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.PostInvitation))))).Methods("POST")

	r.Handle(app.baseURL.Path+"/societies/{society}/invitations/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/societies/{society}/invitations/{id}", app.setSociety(app.verifyToken(app.authenticate(model.AuthAdmin,
		http.HandlerFunc(app.DeleteInvitation))))).Methods("DELETE")

	r.Handle(app.baseURL.Path+"/invitations/{code}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/invitations/{code}", http.HandlerFunc(app.GetInvitationSocietyName)).Methods("GET")
	r.Handle(app.baseURL.Path+"/invitations/{code}", app.verifyToken(http.HandlerFunc(app.AcceptInvitation))).Methods("POST")

	r.Handle(app.baseURL.Path+"/search", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/search", app.verifySearchToken(http.HandlerFunc(app.Search))).Methods("GET")

	r.Handle(app.baseURL.Path+"/search/{id}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/search/{id}", app.verifySearchToken(http.HandlerFunc(app.SearchByID))).Methods("GET")

	r.Handle(app.baseURL.Path+"/search-image/{society}/{id}/{filePath:.*}", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/search-image/{society}/{id}/{filePath:.*}", app.verifySearchToken(http.HandlerFunc(app.SearchImage))).Methods("GET")

	r.Handle(app.baseURL.Path+"/places", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.HandleFunc(app.baseURL.Path+"/places", http.HandlerFunc(app.GetPlacesByPrefix)).Methods("GET")

	r.Handle(app.baseURL.Path+"/current_user", http.HandlerFunc(app.OptionsNoop)).Methods("OPTIONS")
	r.Handle(app.baseURL.Path+"/current_user", app.verifyToken(http.HandlerFunc(app.GetCurrentUser))).Methods("GET")

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
	return getUint32FromRequest(req, "id")
}

func getSocietyIDFromRequest(req *http.Request) (uint32, error) {
	return getUint32FromRequest(req, "society")
}

func getUint32FromRequest(req *http.Request, name string) (uint32, error) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars[name])
	if err != nil || id <= 0 {
		return 0, api.NewError(fmt.Errorf("bad id '%s': %v", vars[name], err))
	}
	return uint32(id), nil
}
