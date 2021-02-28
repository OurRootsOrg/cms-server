package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ourrootsorg/cms-server/model"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/stretchr/testify/assert"
)

func TestSearchByID(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	sh := model.SearchHit{
		ID: "1",
	}
	am.Result = &sh
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/search/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var result model.SearchHit
	err := json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, sh.ID, result.ID)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
}

func TestSearch(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	sr := model.SearchResult{}
	am.Result = &sr
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/search?given=Fred&surname=Flintstone", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t, "Fred", am.Request.(*api.SearchRequest).Given, "Expected given name")
	assert.Equal(t, "Flintstone", am.Request.(*api.SearchRequest).Surname, "Expected surname")
	var result model.SearchResult
	err := json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, result.Total, "Expected empty result, got %#v", result)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
}
