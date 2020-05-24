package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/ourrootsorg/cms-server/api"
)

// PostContentRequest returns a URL for uploading content (via PUT)
// @summary returns a URL for uploading content
// @router /content [post]
// @tags content
// @id postContentRequest
// @Param contentRequest body api.ContentRequest true "Create content request"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PostContentRequest(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	in := api.ContentRequest{}
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	result, errors := app.api.PostContentRequest(req.Context(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(result)
	if err != nil {
		serverError(w, err)
		return
	}
}
