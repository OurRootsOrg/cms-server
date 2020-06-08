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
