package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
)

func TestCollections(t *testing.T) {
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
	app := NewApp().
		CategoryPersister(persist.NewPostgresPersister("", db)).
		CollectionPersister(persist.NewPostgresPersister("", db))

	r := app.NewRouter()

	// Add a test category for referential integrity
	testCategory, err := createTestCategory(r)
	assert.Nil(t, err, "Error creating test category")
	defer deleteTestCategory(r, testCategory)

	request, _ := http.NewRequest("GET", "/collections", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Response: %s", string(response.Body.Bytes()))
	var empty []model.Collection
	err = json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Add a Collection
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{
			Name: "Test Collection",
		},
		Category: testCategory.CategoryRef,
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
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "Response: %s", string(response.Body.Bytes()))
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
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "Response: %s", string(response.Body.Bytes()))
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
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
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

	// GET /collections should now return the created Collection
	request, _ = http.NewRequest("GET", "/collections", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Response: %s", string(response.Body.Bytes()))
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
	assert.Equal(t, 200, response.Code, "Response: %s", string(response.Body.Bytes()))
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
	assert.Equal(t, http.StatusBadRequest, response.Code, "Response: %s", string(response.Body.Bytes()))

	// Bad request - no category
	in.Category = model.CategoryRef{}
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	request, _ = http.NewRequest("POST", "/collections", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code, "Response: %s", string(response.Body.Bytes()))
	// Collection not found
	request, _ = http.NewRequest("GET", created.ID+"999", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code, "Response: %s", string(response.Body.Bytes()))

	// PUT
	ret2.Name = "Updated"
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(ret2)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	// correct MIME type
	request, _ = http.NewRequest("PUT", created.ID, buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var updated model.Collection
	err = json.NewDecoder(response.Body).Decode(&updated)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, ret2.Name, updated.Name, "Expected Name to match")

	// Missing MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	request, _ = http.NewRequest("PUT", created.ID, buf)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "Response: %s", string(response.Body.Bytes()))
	// Bad MIME type
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	request, _ = http.NewRequest("PUT", created.ID, buf)
	request.Header.Add("Content-Type", "application/notjson")
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnsupportedMediaType, response.Code, "Response: %s", string(response.Body.Bytes()))

	// PUT non-existant
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(ret2)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	request, _ = http.NewRequest("PUT", created.ID+"x", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code, "Response: %s", string(response.Body.Bytes()))

	// Bad request
	request, _ = http.NewRequest("PUT", created.ID, strings.NewReader("{x}"))
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

func createTestCategory(r *mux.Router) (*model.Category, error) {
	stringType, err := model.NewFieldDef("stringField", model.StringType, "string_field")
	if err != nil {
		return nil, err
	}
	in, err := model.NewCategoryIn("Test", stringType)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}
	request, _ := http.NewRequest("POST", "/categories", buf)
	request.Header.Add("Content-Type", contentType)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		return nil, errors.New("Error creating category")
	}
	var created model.Category
	err = json.NewDecoder(response.Body).Decode(&created)
	return &created, err
}

func deleteTestCategory(r *mux.Router, category *model.Category) {
	request, _ := http.NewRequest("DELETE", category.ID, nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
}
