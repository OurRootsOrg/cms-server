package api

import (
	"bytes"
	"context"
	"encoding/json"
	"gocloud.dev/postgres"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
)

func TestCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	db, err := postgres.Open(context.TODO(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
			err,
			os.Getenv("DATABASE_URL"),
		)
	}
	app := NewApp().CategoryPersister(persist.NewPostgresPersister("", db))
	r := app.NewRouter()

	request, _ := http.NewRequest("GET", "/categories", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty []model.Category
	err = json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	// Add a Category
	stringType, err := model.NewFieldDef("stringField", model.StringType, "string_field")
	assert.NoError(t, err)
	in, err := model.NewCategoryIn("First", stringType)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	// missing MIME type
	request, _ = http.NewRequest("POST", "/categories", buf)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "415 response is expected")
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	// wrong MIME type
	request, _ = http.NewRequest("POST", "/categories", buf)
	request.Header.Add("Content-Type", "application/notjson")
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	// correct MIME type
	request, _ = http.NewRequest("POST", "/categories", buf)
	request.Header.Add("Content-Type", contentType)

	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Category
	err = json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.Equal(t, in.FieldDefs, created.FieldDefs, "Expected FieldDefs to match")
	assert.NotEmpty(t, created.ID)
	// t.Logf("app.Categories: %#v", app.Categories)
	// GET /categories should now return the created Category
	request, _ = http.NewRequest("GET", "/categories", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret []model.Category
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret))
	assert.Equal(t, created, ret[0])
	// GET /categories/{id} should now return the created Category
	request, _ = http.NewRequest("GET", created.ID, nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret2 model.Category
	err = json.NewDecoder(response.Body).Decode(&ret2)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, created, ret2)
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Bad request
	request, _ = http.NewRequest("POST", "/categories", strings.NewReader("{xxx}"))
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code, "400 response is expected")

	// Category not found
	request, _ = http.NewRequest("GET", created.ID+"999", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code, "404 response is expected")

	// PATCH
	in.Name = "Updated"
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	// correct MIME type
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var updated model.Category
	err = json.NewDecoder(response.Body).Decode(&updated)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, updated.Name, "Expected Name to match")
	assert.Equal(t, in.FieldDefs, updated.FieldDefs, "Expected FieldDefs to match")

	// Missing MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "Response: %s", string(response.Body.Bytes()))
	// Bad MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	request.Header.Add("Content-Type", "application/notjson")
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "Response: %s", string(response.Body.Bytes()))

	// PATCH non-existant
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID+"999", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code, "Response: %s", string(response.Body.Bytes()))

	// Bad request
	request, _ = http.NewRequest("PATCH", created.ID, strings.NewReader("{x}"))
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code, "Response: %s", string(response.Body.Bytes()))

	// DELETE
	request, _ = http.NewRequest("DELETE", created.ID, nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code, "Response: %s", string(response.Body.Bytes()))
}
