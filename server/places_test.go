package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestGetPlacesByPrefix(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	res := []model.Place{}
	am.Result = res
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/places?prefix=Ala&count=20", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t, "Ala", am.Request.(string), "Expected Ala prefix")
	var result []model.Place
	err := json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(result), "Expected empty result, got %#v", result)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
}
