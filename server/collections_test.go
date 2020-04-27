package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/jancona/ourroots/persist"
	"github.com/stretchr/testify/assert"
)

func TestCollections(t *testing.T) {
	app := App{
		CollectionPersister: persist.NewMemoryPersister(""),
	}
	r := NewRouter(app)

	request, _ := http.NewRequest("GET", "/collections", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty []model.Collection
	err := json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	// Add a Collection
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{},
		Category:       model.NewCategoryRef(1),
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	// missing MIME type
	request, _ = http.NewRequest("POST", "/collections", buf)
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
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	// wrong MIME type
	request, _ = http.NewRequest("POST", "/collections", buf)
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
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	// correct MIME type
	request, _ = http.NewRequest("POST", "/collections", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Collection
	err = json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Category, created.Category)
	// t.Logf("app.Collections: %#v", app.Collections)
	// GET /collections should now return the created Collection
	request, _ = http.NewRequest("GET", "/collections", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret []model.Collection
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret))
	assert.Equal(t, created, ret[0])
	// GET /collections/{id} should now return the created Collection
	request, _ = http.NewRequest("GET", created.ID, nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret2 model.Collection
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
	request, _ = http.NewRequest("POST", "/collections", strings.NewReader("{xxx}"))
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code, "400 response is expected")

	// Collection not found
	request, _ = http.NewRequest("GET", created.ID+"x", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code, "404 response is expected")

	// PATCH
	in.Name = "Updated"
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
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
	var updated model.Collection
	err = json.NewDecoder(response.Body).Decode(&updated)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, updated.Name, "Expected Name to match")

	// Missing MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code)
	// Bad MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	request, _ = http.NewRequest("PATCH", created.ID, buf)
	request.Header.Add("Content-Type", "application/notjson")
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code)

	// PATCH non-existant
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
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
