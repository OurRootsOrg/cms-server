package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// GetRecords returns all records for a post
// @summary returns all records
// @router /records [get]
// @tags records
// @id getRecords
// @produce application/json
// @success 200 {array} api.RecordsResult "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetRecords(w http.ResponseWriter, req *http.Request) {
	postStr := req.URL.Query().Get("post")
	postID, err := strconv.Atoi(postStr)
	if err != nil || postStr == "" || postID <= 0 {
		ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Missing or invalid post query parameter '%s'", postStr))
		return
	}
	log.Printf("[DEBUG] Get records for post ID: %d", postID)
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cols, errors := app.api.GetRecordsForPost(req.Context(), uint32(postID))
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err = enc.Encode(cols)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetRecord gets a Record from the database
// @summary gets a Record with optional detail including household records and image path
// @router /records/{id} [get]
// @tags posts
// @id getRecord
// @Param id path integer true "Record ID"
// @produce application/json
// @param details query bool false "include labels, citation, household, and imagePath"
// @success 200 {object} api.RecordDetail "OK"
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetRecord(w http.ResponseWriter, req *http.Request) {
	recordID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	details, _ := strconv.ParseBool(req.URL.Query().Get("details"))
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	record, errors := app.api.GetRecord(req.Context(), details, recordID)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(record)
	if err != nil {
		serverError(w, err)
		return
	}
}
