package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// GetCollections returns all collections in the database
// @summary returns all collections
// @router /collections [get]
// @tags collections
// @id getCollections
// @produce application/json
// @success 200 {array} model.Collection "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetCollections(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cols, errors := app.api.GetCollections(app.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(cols)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetCollection gets a Collection from the database
// @summary gets a Collection
// @router /collections/{id} [get]
// @tags collections
// @id getCollection
// @Param id path string true "Collection ID" format(url)
// @produce application/json
// @success 200 {object} model.Collection "OK"
// @failure 404 {object} model.Errors "Not found"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetCollection(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	collection, errors := app.api.GetCollection(app.Context(), req.URL.String())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(collection)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PostCollection adds a new Collection to the database
// @summary adds a new Collection
// @router /collections [post]
// @tags collections
// @id addCollection
// @Param collection body model.CollectionIn true "Add Collection"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Collection "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PostCollection(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	in := model.CollectionIn{}
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err)
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	collection, errors := app.api.AddCollection(app.Context(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(collection)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PutCollection updates a Collection in the database
// @summary updates a Collection
// @router /collections/{id} [put]
// @tags collections
// @id updateCollection
// @Param id path string true "Collection ID" format(url)
// @Param collection body model.Collection true "Update Collection"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Collection "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PutCollection(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	var in model.Collection
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err)
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	collection, errors := app.api.UpdateCollection(app.Context(), req.URL.String(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err = enc.Encode(collection)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeleteCollection deletes a Collection from the database
// @summary deletes a Collection
// @router /collections/{id} [delete]
// @tags collections
// @id deleteCollection
// @Param id path string true "Collection ID" format(url)
// @success 204 {object} model.Collection "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) DeleteCollection(w http.ResponseWriter, req *http.Request) {
	errors := app.api.DeleteCollection(app.Context(), req.URL.String())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
