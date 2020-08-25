package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// GetPlacesByPrefix returns places matching prefix
// @summary returns places matching prefix
// @router /places [get]
// @param prefix query string false "place prefix"
// @param count query int false "maximum number of places to return"
// @tags places
// @id getPlacesByPrefix
// @produce application/json
// @success 200 {array} model.Place "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetPlacesByPrefix(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "max-age=7200")
	prefix := req.URL.Query().Get("prefix")
	count, err := strconv.Atoi(req.URL.Query().Get("count"))
	if err != nil || count <= 0 || count > 20 {
		count = 10
	}
	places, errors := app.api.GetPlacesByPrefix(req.Context(), prefix, count)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err = enc.Encode(places)
	if err != nil {
		serverError(w, err)
		return
	}
}
