package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"

	"github.com/jancona/ourroots/model"
	"github.com/jancona/ourroots/persist"
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
	cols, err := app.CollectionPersister.SelectCollections()
	if err != nil {
		serverError(w, err)
		return
	}
	v := make([]model.Collection, 0, len(cols))
	for _, value := range cols {
		v = append(v, value)
	}
	err = enc.Encode(v)
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
	in := model.CollectionIn{}
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	collection, err := app.CollectionPersister.InsertCollection(in)
	if err == persist.ErrForeignKeyViolation {
		msg := fmt.Sprintf("Invalid category reference: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	} else if err != nil {
		serverError(w, err)
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
	collection, err := app.CollectionPersister.SelectOneCollection(req.URL.String())
	if err == persist.ErrNoRows {
		notFound(w, req)
		return
	} else if err != nil {
		serverError(w, err)
		return
	}
	err = enc.Encode(collection)
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
// @Param collection body model.CollectionIn true "Update Collection"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Collection "OK"
// @failure 415 {object} model.Error "Bad Content-Type"
// @failure 500 {object} model.Error "Server error"
func (app App) PatchCollection(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	var in model.CollectionIn
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	collection, err := app.CollectionPersister.UpdateCollection(req.URL.String(), in)
	if err == persist.ErrNoRows {
		// Not allowed to add a Collection with PATCH
		notFound(w, req)
		return
	} else if err != nil {
		serverError(w, err)
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
// @Param id path string true "Collection ID" format(uuid)
// @success 204 {object} model.Collection "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) DeleteCollection(w http.ResponseWriter, req *http.Request) {
	err := app.CollectionPersister.DeleteCollection(req.URL.String())
	if err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
