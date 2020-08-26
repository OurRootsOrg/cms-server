package main

import (
	"encoding/json"
	"net/http"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
)

// GetCurrentUser returns the current user
// @summary returns the current user
// @router /currentuser [get]
// @tags users
// @id getCurrentUser
// @produce application/json
// @success 200 {array} model.Place "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetCurrentUser(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	user := req.Context().Value(api.UserProperty).(*model.User)
	err := enc.Encode(user)
	if err != nil {
		serverError(w, err)
		return
	}
}
