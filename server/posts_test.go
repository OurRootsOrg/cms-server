package main

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

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
)

func TestPosts(t *testing.T) {
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
	p := persist.NewPostgresPersister("", db)
	app := NewApp().
		API(api.NewAPI().
			CategoryPersister(p).
			CollectionPersister(p).
			PostPersister(p))

	r := app.NewRouter()

	// Add a test collection for referential integrity
	testCategory, err := createTestCategory(r)
	assert.Nil(t, err, "Error creating test category")
	defer deleteTestCategory(r, testCategory)
	testCollection, err := createTestCollection(r, testCategory)
	assert.Nil(t, err, "Error creating test collection")
	defer deleteTestCollection(r, testCollection)

	request, _ := http.NewRequest("GET", "/posts", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Response: %s", string(response.Body.Bytes()))
	var empty api.PostResult
	err = json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty.Posts), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Add a Post
	in := model.PostIn{
		PostBody: model.PostBody{
			Name: "Test Post",
		},
		Collection: testCollection.ID,
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding PostIn: %v", err)
	}
	// missing MIME type
	request, _ = http.NewRequest("POST", "/posts", buf)
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
		t.Errorf("Error encoding PostIn: %v", err)
	}
	// wrong MIME type
	request, _ = http.NewRequest("POST", "/posts", buf)
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
		t.Errorf("Error encoding PostIn: %v", err)
	}
	// correct MIME type
	request, _ = http.NewRequest("POST", "/posts", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Post
	err = json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Collection, created.Collection, "created: %#v", created)

	// GET /posts should now return the created Post
	request, _ = http.NewRequest("GET", "/posts", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret api.PostResult
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret.Posts))
	assert.Equal(t, created, ret.Posts[0])

	// GET /posts/{id} should now return the created Post
	request, _ = http.NewRequest("GET", created.ID, nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret2 model.Post
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
	request, _ = http.NewRequest("POST", "/posts", strings.NewReader("{xxx}"))
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code, "Response: %s", string(response.Body.Bytes()))

	// Bad request - no collection
	in.Collection = ""
	buf = new(bytes.Buffer)
	enc = json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding PostIn: %v", err)
	}
	request, _ = http.NewRequest("POST", "/posts", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code, "Response: %s", string(response.Body.Bytes()))
	// Post not found
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
		t.Errorf("Error encoding PostIn: %v", err)
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
	var updated model.Post
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
		t.Errorf("Error encoding PostIn: %v", err)
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
		t.Errorf("Error encoding PostIn: %v", err)
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
		t.Errorf("Error encoding PostIn: %v", err)
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

func createTestCollection(r *mux.Router, category *model.Category) (*model.Collection, error) {
	in := model.NewCollectionIn("Test", category.ID)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}
	request, _ := http.NewRequest("POST", "/collections", buf)
	request.Header.Add("Content-Type", contentType)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		return nil, errors.New("Error creating collection")
	}
	var created model.Collection
	err := json.NewDecoder(response.Body).Decode(&created)
	return &created, err
}

func deleteTestCollection(r *mux.Router, collection *model.Collection) {
	request, _ := http.NewRequest("DELETE", collection.ID, nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
}
