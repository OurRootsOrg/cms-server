package main

import (
	"encoding/json"
	"fmt"
	"log"
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
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetCollections(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cols, errors := app.api.GetCollections(req.Context())
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
// @Param id path integer true "Collection ID"
// @produce application/json
// @success 200 {object} model.Collection "OK"
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetCollection(w http.ResponseWriter, req *http.Request) {
	collID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	collection, errors := app.api.GetCollection(req.Context(), collID)
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
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
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
	collection, errors := app.api.AddCollection(req.Context(), in)
	if errors != nil {
		log.Printf("[DEBUG] PostCollection AddCollection %v\n", errors)
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
// @Param id path integer true "Collection ID"
// @Param collection body model.Collection true "Update Collection"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Collection "OK"
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PutCollection(w http.ResponseWriter, req *http.Request) {
	collID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
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
	collection, errors := app.api.UpdateCollection(req.Context(), collID, in)
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
// @Param id path integer true "Collection ID"
// @success 204 {object} model.Collection "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) DeleteCollection(w http.ResponseWriter, req *http.Request) {
	collID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	errors = app.api.DeleteCollection(req.Context(), collID)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
