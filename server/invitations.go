package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ourrootsorg/cms-server/model"
)

// GetInvitations returns all invitations for a society
// @summary returns all invitations
// @router /societies/{society}/invitations [get]
// @tags invitations
// @id getInvitations
// @produce application/json
// @success 200 {array} model.Invitation "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetInvitations(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	societies, errors := app.api.GetInvitations(req.Context())
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

// PostInvitation adds a new Invitation to the database
// @summary adds a new Invitation
// @router /societies/{society}/invitations [post]
// @tags invitations
// @id addInvitation
// @Param society body model.InvitationIn true "Add Invitation"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Invitation "OK"
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PostInvitation(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	body := model.InvitationBody{}
	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	invitation, errors := app.api.AddInvitation(req.Context(), body)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(invitation)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeleteInvitation deletes an Invitation from the database
// @summary deletes an Invitation
// @router /societies/{society}/invitations/{id} [delete]
// @tags invitations
// @id deleteInvitation
// @Param id path integer true "Invitation ID"
// @success 204 "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) DeleteInvitation(w http.ResponseWriter, req *http.Request) {
	invID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	errors = app.api.DeleteInvitation(req.Context(), invID)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetInvitation gets an Invitation from the database by the invitation code
// @summary gets an Invitation
// @router /invitations/{code} [get]
// @tags invitations
// @id getInvitation
// @produce application/json
// @success 200 {object} api.InvitationSocietyName "OK"
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetInvitationSocietyName(w http.ResponseWriter, req *http.Request) {
	code := mux.Vars(req)["code"]
	if code == "" {
		msg := fmt.Sprintf("Bad request: missing code query parameter")
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	society, errors := app.api.GetInvitationSocietyName(req.Context(), code)
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

// AcceptInvitation accepts an invitation
// @summary accept an invitation
// @router /invitations/{code} [post]
// @tags invitations
// @id acceptInvitation
// @accept application/json
// @produce application/json
// @success 200 {object} model.SocietyUser "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) AcceptInvitation(w http.ResponseWriter, req *http.Request) {
	code := mux.Vars(req)["code"]
	societyUser, errors := app.api.AcceptInvitation(req.Context(), code)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err := enc.Encode(societyUser)
	if err != nil {
		serverError(w, err)
		return
	}
}
