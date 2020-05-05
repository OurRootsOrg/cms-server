package api

import (
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ourrootsorg/cms-server/model"
)

// API is the container for the apilication
type API struct {
	categoryPersister   model.CategoryPersister
	collectionPersister model.CollectionPersister
	baseURL             url.URL
	validate            *validator.Validate
}

// NewAPI builds an API
func NewAPI() *API {
	api := &API{
		baseURL: url.URL{},
	}
	api.Validate(validator.New())
	return api
}

// BaseURL sets the base URL for the api
func (api *API) BaseURL(url url.URL) *API {
	api.baseURL = url
	return api
}

// Validate sets the validate object for the api
func (api *API) Validate(validate *validator.Validate) *API {
	// Return JSON tag name as Field() in errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	api.validate = validate
	return api
}

// CategoryPersister sets the CategoryPersister for the api
func (api *API) CategoryPersister(cp model.CategoryPersister) *API {
	api.categoryPersister = cp
	return api
}

// CollectionPersister sets the CollectionPersister for the api
func (api *API) CollectionPersister(cp model.CollectionPersister) *API {
	api.collectionPersister = cp
	return api
}
