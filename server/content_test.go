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

func TestContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	app := NewApp().
		API(api.NewAPI().
			BlobStoreConfig("us-east-1", "127.0.0.1:19000",
				"minioaccess", "miniosecret", "testbucket", true))
	r := app.NewRouter()

	request, _ := http.NewRequest("POST", "/content", strings.NewReader("{\"contentType\": \"text/csv\"}"))
	request.Header.Add("Content-Type", contentType)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var result api.ContentResult
	err := json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotEmpty(t, result.Key)
	assert.NotEmpty(t, result.PutURL)
}
