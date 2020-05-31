package main

import (
	"encoding/json"
	"net/http"
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
	postID := req.URL.Query().Get("post")
	if postID == "" {
		ErrorResponse(w, http.StatusBadRequest, "post query parameter required")
		return
	}
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cols, errors := app.api.GetRecordsForPost(req.Context(), postID)
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
