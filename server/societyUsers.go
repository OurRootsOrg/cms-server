package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/ourrootsorg/cms-server/api"
)

// GetUsers returns all users for a society
// @summary returns all users
// @router /societies/{society}/users [get]
// @tags societyUsers
// @id getUsers
// @produce application/json
// @success 200 {array} api.SocietyUserName "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetSocietyUserNames(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	societies, errors := app.api.GetSocietyUserNames(req.Context())
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

// GetCurrentSocietyUser returns the current SocietyUser
// @summary returns the current SocietyUser
// @router /societies/{society}/current_user [get]
// @tags societyUsers
// @id getCurrentSocietyUser
// @produce application/json
// @success 200 {array} model.SocietyUser "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetCurrentSocietyUser(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	user, err := utils.GetUserFromContext(req.Context())
	if err != nil {
		ErrorsResponse(w, err)
		return
	}
	societyUser, errors := app.api.GetSocietyUserByUser(req.Context(), user.ID)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err = enc.Encode(societyUser)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PutSocietyUserName updates a SocietyUser in the database
// @summary updates a SocietyUser
// @router /societies/{society}/users/{id} [put]
// @tags societyUsers
// @id updateSocietyUser
// @Param id path integer true "SocietyUser ID"
// @Param society body api.SocietyUserName true "Update SocietyUser"
// @accept application/json
// @produce application/json
// @success 200 {object} api.SocietyUserName "OK"
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PutSocietyUserName(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	var in api.SocietyUserName
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	society, errors := app.api.UpdateSocietyUserName(req.Context(), in.ID, in)
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

// DeleteSociety deletes a SocietyUser from the database
// @summary deletes a SocietyUser
// @router /societies/{society}/users/{id} [delete]
// @tags societyUsers
// @id deleteSocietyUSer
// @Param id path integer true "SocietyUser ID"
// @success 204 "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) DeleteSocietyUser(w http.ResponseWriter, req *http.Request) {
	id, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	errors = app.api.DeleteSocietyUser(req.Context(), id)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
