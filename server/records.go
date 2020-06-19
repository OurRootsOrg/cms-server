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
// @success 200 {array} model.Record "OK"
// @failure 500 {object} model.Errors "Server error"
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
