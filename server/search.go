package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ourrootsorg/cms-server/api"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

// Search returns search results matching a query
// @summary returns search results
// @router /search [get]
// @tags search
// @id search
// @produce application/json
// @success 200 {array} model.SearchResult "OK"
// @failure 500 {object} api.Errors "Server error"
// TODO need to specify possible query parameters
func (app App) Search(w http.ResponseWriter, req *http.Request) {
	var searchRequest api.SearchRequest
	err := decoder.Decode(&searchRequest, req.URL.Query())
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}

	result, errors := app.api.Search(req.Context(), &searchRequest)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	err = enc.Encode(result)
	if err != nil {
		serverError(w, err)
		return
	}
}

// SearchByID returns detailed information about a single search result
// @summary returns a single search result
// @router /search/{id} [get]
// @tags search
// @id searchByID
// @Param id path string true "Search Result ID"
// @produce application/json
// @success 200 {object} model.SearchHit "OK"
// @failure 404 {object} api.Errors "Not found"
// @failure 500 {object} api.Errors "Server error"
func (app App) SearchByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	result, errors := app.api.SearchByID(req.Context(), vars["id"])
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	err := enc.Encode(result)
	if err != nil {
		serverError(w, err)
		return
	}
}
