package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// GetSettings returns the global settings object
// @summary returns goobal settings
// @router /posts [get]
// @tags settings
// @id getSettings
// @produce application/json
// @success 200 {object} model.Settings "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetSettings(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	settings, errors := app.api.GetSettings(req.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(settings)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PutSettings updates the global settings object
// @summary updates settings
// @router /settings [put]
// @tags settings
// @id updateSettings
// @Param post body model.Settings true "Update Settings"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Settings "OK"
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PutSettings(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	var in model.Settings
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err)
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	settings, errors := app.api.UpdateSettings(req.Context(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err = enc.Encode(settings)
	if err != nil {
		serverError(w, err)
		return
	}
}
