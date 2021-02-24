package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// GetSocietySummaries returns all society summaries for the current user
// @summary returns all society summaries for the current user
// @router /society_summaries [get]
// @tags societies
// @id getSocietySummaries
// @produce application/json
// @success 200 {array} model.SocietySummary "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetSocietySummariesForCurrentUser(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	societies, errors := app.api.GetSocietySummariesForCurrentUser(req.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(societies)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetSocietySummary gets a SocietySummary from the database
// @summary gets a SocietySummary
// @router /society_summaries/{society} [get]
// @tags societies
// @id getSocietySummary
// @Param id path integer true "Society ID"
// @produce application/json
// @success 200 {object} model.SocietySummary "OK"
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetSocietySummary(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	society, errors := app.api.GetSocietySummary(req.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(society)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PostSociety adds a new Society to the database
// @summary adds a new Society
// @router /societies [post]
// @tags societies
// @id addSociety
// @Param society body model.SocietyIn true "Add Society"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Society "OK"
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PostSociety(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	in := model.SocietyIn{}
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	society, errors := app.api.AddSociety(req.Context(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(society)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetSociety gets a Society from the database
// @summary gets a Society
// @router /societies/{society} [get]
// @tags societies
// @id getSociety
// @Param id path integer true "Society ID"
// @produce application/json
// @success 200 {object} model.Society "OK"
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetSociety(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	society, errors := app.api.GetSociety(req.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(society)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PutSociety updates a Society in the database
// @summary updates a Society
// @router /societies/{id} [put]
// @tags societies
// @id updateSociety
// @Param id path integer true "Society ID"
// @Param society body model.Society true "Update Society"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Society "OK"
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PutSociety(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	var in model.Society
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	society, errors := app.api.UpdateSociety(req.Context(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err = enc.Encode(society)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeleteSociety deletes a Society from the database
// @summary deletes a Society
// @router /societies/{id} [delete]
// @tags societies
// @id deleteSociety
// @Param id path integer true "Society ID"
// @success 204 "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) DeleteSociety(w http.ResponseWriter, req *http.Request) {
	errors := app.api.DeleteSociety(req.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
