package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestCategories(t *testing.T) {
	app := App{
		Categories: make(map[string]model.Category),
	}
	r := NewRouter(app)

	request, _ := http.NewRequest("GET", "/categories", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty []model.Category
	err := json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	// Add a Category
	stringType, err := model.NewFieldDef("stringField", model.StringType)
	assert.NoError(t, err)
	ci, err := model.NewCategoryInput("First", stringType)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(ci)
	if err != nil {
		t.Errorf("Error encoding CategoryInput: %v", err)
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
	err = enc.Encode(ci)
	if err != nil {
		t.Errorf("Error encoding CategoryInput: %v", err)
	}
	// wrong MIME type
	request, _ = http.NewRequest("POST", "/categories", buf)
	request.Header.Add("Content-Type", "application/notjson")
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code)
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(ci)
	if err != nil {
		t.Errorf("Error encoding CategoryInput: %v", err)
	}
	// correct MIME type
	request, _ = http.NewRequest("POST", "/categories", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Category
	err = json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, ci.Name, created.Name, "Expected Name to match")
	assert.Equal(t, ci.FieldDefs, created.FieldDefs, "Expected FieldDefs to match")
	assert.Equal(t, ci.CSVHeading, created.CSVHeading, "Expected CSVHeading to match")
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
	request, _ = http.NewRequest("GET", created.ID+"x", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code, "404 response is expected")

	// PATCH
	ci.Name = "Updated"
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(ci)
	if err != nil {
		t.Errorf("Error encoding CategoryInput: %v", err)
	}
	// correct MIME type
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var updated model.Category
	err = json.NewDecoder(response.Body).Decode(&updated)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, ci.Name, updated.Name, "Expected Name to match")
	assert.Equal(t, ci.FieldDefs, updated.FieldDefs, "Expected FieldDefs to match")
	assert.Equal(t, ci.CSVHeading, updated.CSVHeading, "Expected CSVHeading to match")

	// Missing MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(ci)
	if err != nil {
		t.Errorf("Error encoding CategoryInput: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code)
	// Bad MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(ci)
	if err != nil {
		t.Errorf("Error encoding CategoryInput: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	request.Header.Add("Content-Type", "application/notjson")
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code)

	// PATCH non-existant
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(ci)
	if err != nil {
		t.Errorf("Error encoding CategoryInput: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID+"x", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)

	// Bad request
	request, _ = http.NewRequest("PATCH", created.ID, strings.NewReader("{x}"))
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// DELETE
	request, _ = http.NewRequest("DELETE", created.ID, nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code)
}
