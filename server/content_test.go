package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/stretchr/testify/assert"
)

func TestPostContentRequest(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	r := app.NewRouter()

	am.result = &api.ContentResult{
		Key:    "path/key",
		PutURL: "https://s3.example.com/bucket/path/key",
	}
	am.errors = nil

	request, _ := http.NewRequest("POST", "/content", strings.NewReader("{\"contentType\": \"text/csv\"}"))
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var result api.ContentResult
	err := json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotEmpty(t, result.Key)
	assert.NotEmpty(t, result.PutURL)
}
