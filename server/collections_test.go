package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestGetAllCollections(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.CollectionResult{}
	am.result = &cr
	am.errors = nil

	request, _ := http.NewRequest("GET", "/collections", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty api.CollectionResult
	err := json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty.Collections), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Non-empty result
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	ci, _ := makeCollectionIn(t)
	cr = api.CollectionResult{
		Collections: []model.Collection{
			{
				ID:             "/collections/1",
				CollectionIn:   ci,
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.result = &cr
	am.errors = nil
	request, _ = http.NewRequest("GET", "/collections", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret api.CollectionResult
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret.Collections))
	assert.Equal(t, cr.Collections[0], ret.Collections[0])

	// error result
	am.result = (*api.CollectionResult)(nil)
	am.errors = model.NewErrors(http.StatusInternalServerError, assert.AnError)
	request, _ = http.NewRequest("GET", "/collections", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 500, response.Code)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var errRet []model.Error
	err = json.NewDecoder(response.Body).Decode(&errRet)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotNil(t, errRet)
	assert.Equal(t, 1, len(errRet))
	assert.Equal(t, am.errors.Errs(), errRet)
}
func TestGetCollection(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	ci, _ := makeCollectionIn(t)
	collection := &model.Collection{
		ID:           "/collections/1",
		CollectionIn: ci,
	}
	am.result = collection
	am.errors = nil
	var ret model.Collection

	request, _ := http.NewRequest("GET", "/collections/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	err := json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, *collection, ret)

	collection = nil
	am.result = collection
	am.errors = model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, "/collections/1"))

	request, _ = http.NewRequest("GET", "/collections/1", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var errRet []model.Error
	err = json.NewDecoder(response.Body).Decode(&errRet)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotNil(t, errRet)
	assert.Equal(t, 1, len(errRet))
	assert.Equal(t, am.errors.Errs(), errRet)
}

func TestPostCollection(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makeCollectionIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	am.result = &model.Collection{
		ID:             "/collections/1",
		CollectionIn:   in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.errors = nil

	request, _ := http.NewRequest("POST", "/collections", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Collection
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.Equal(t, in, created.CollectionIn)
	assert.Equal(t, now, created.InsertTime)
	assert.Equal(t, now, created.LastUpdateTime)
	assert.NotEmpty(t, created.ID)
}

func TestPutCollection(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makeCollectionIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	coll := model.Collection{
		ID:             "/collections/1",
		CollectionIn:   in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.result = &coll
	am.errors = nil

	request, _ := http.NewRequest("PUT", "/collections/1", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Collection
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.Equal(t, in, created.CollectionIn)
	assert.Equal(t, now, created.InsertTime)
	assert.Equal(t, now, created.LastUpdateTime)
	assert.NotEmpty(t, created.ID)
}

func TestDeleteCollection(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	am.result = nil
	am.errors = nil

	request, _ := http.NewRequest("DELETE", "/collections/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code, "Response: %s", string(response.Body.Bytes()))
}

func makeCollectionIn(t *testing.T) (model.CollectionIn, *bytes.Buffer) {
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{
			Name: "First",
		},
		Category: "/categories/1",
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	return in, buf
}
