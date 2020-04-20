package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"

	"github.com/jancona/ourroots/model"
)

// GetAllCollections returns all collections in the database
// @summary returns all collections
// @router /collections [get]
// @tags collections
// @id getCollections
// @produce application/json
// @success 200 {array} model.Collection "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) GetAllCollections(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	v := make([]model.Collection, 0, len(app.Collections))

	for _, value := range app.Collections {
		v = append(v, value)
	}
	err := enc.Encode(v)
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
// @Param collection body model.CollectionInput true "Add Collection"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Collection "OK"
// @failure 415 {object} model.Error "Bad Content-Type"
// @failure 500 {object} model.Error "Server error"
func (app App) PostCollection(w http.ResponseWriter, req *http.Request) {
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
	ci := model.CollectionInput{Name: ""}
	err = json.NewDecoder(req.Body).Decode(&ci)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	collection := model.NewCollection(ci)
	log.Printf("collection: %#v", collection)
	// Add to "database"
	app.Collections[collection.ID] = collection
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(collection)
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
// @Param id path string true "Collection ID" format(uuid)
// @produce application/json
// @success 200 {object} model.Collection "OK"
// @failure 404 {object} model.Error "Not found"
// @failure 500 {object} model.Error "Server error"
func (app App) GetCollection(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	collection, found := app.Collections[req.URL.String()]
	if !found {
		notFound(w, req)
		return
	}
	err := enc.Encode(collection)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PatchCollection updates a Collection in the database
// @summary updates a Collection
// @router /collections/{id} [patch]
// @tags collections
// @id updateCollection
// @Param id path string true "Collection ID" format(uuid)
// @Param collection body model.CollectionInput true "Update Collection"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Collection "OK"
// @failure 415 {object} model.Error "Bad Content-Type"
// @failure 500 {object} model.Error "Server error"
func (app App) PatchCollection(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		serverError(w, err)
		return
	}
	_, found := app.Collections[req.URL.String()]
	if !found {
		// Not allowed to add a Collection with PATCH
		notFound(w, req)
		return
	}
	if mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	var tdi model.CollectionInput
	err = json.NewDecoder(req.Body).Decode(&tdi)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	collection := app.Collections[req.URL.String()]
	collection.Name = tdi.Name
	// Add to "database"
	app.Collections[req.URL.String()] = collection
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
// @Param id path string true "Collection ID" format(uuid)
// @success 204 {object} model.Collection "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) DeleteCollection(w http.ResponseWriter, req *http.Request) {
	delete(app.Collections, req.URL.String())
	w.WriteHeader(http.StatusNoContent)
}
