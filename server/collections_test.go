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
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.CollectionResult{}
	am.Result = &cr
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/societies/1/collections", nil)
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
	ci, _ := makeCollectionIn(t, 1)
	cr = api.CollectionResult{
		Collections: []model.Collection{
			{
				ID:             1,
				CollectionIn:   ci,
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.Result = &cr
	am.Errors = nil
	request, _ = http.NewRequest("GET", "/societies/1/collections", nil)
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
	am.Result = (*api.CollectionResult)(nil)
	am.Errors = api.NewError(assert.AnError)
	request, _ = http.NewRequest("GET", "/societies/1/collections", nil)
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
	assert.Equal(t, am.Errors.(*api.Error).Errs(), errRet)
}
func TestGetCollection(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	ci, _ := makeCollectionIn(t, 1)
	collection := &model.Collection{
		ID:           1,
		CollectionIn: ci,
	}
	am.Result = collection
	am.Errors = nil
	var ret model.Collection

	request, _ := http.NewRequest("GET", "/societies/1/collections/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	err := json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, *collection, ret)

	collection = nil
	am.Result = collection
	am.Errors = api.NewError(model.NewError(model.ErrNotFound, "1"))

	request, _ = http.NewRequest("GET", "/societies/1/collections/1", nil)
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
	assert.Equal(t, am.Errors.(*api.Error).Errs(), errRet)
}

func TestPostCollection(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makeCollectionIn(t, 1)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	am.Result = &model.Collection{
		ID:             1,
		CollectionIn:   in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.Errors = nil

	request, _ := http.NewRequest("POST", "/societies/1/collections", buf)
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
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makeCollectionIn(t, 1)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	coll := model.Collection{
		ID:             1,
		CollectionIn:   in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.Result = &coll
	am.Errors = nil

	request, _ := http.NewRequest("PUT", "/societies/1/collections/1", buf)
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
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	am.Result = nil
	am.Errors = nil

	request, _ := http.NewRequest("DELETE", "/societies/1/collections/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code, "Response: %s", string(response.Body.Bytes()))
}

func makeCollectionIn(t *testing.T, categoryID uint32) (model.CollectionIn, *bytes.Buffer) {
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{
			Name:           "First",
			CollectionType: model.CollectionTypeRecords,
		},
		Categories: []uint32{},
	}
	if categoryID > 0 {
		in.Categories = []uint32{categoryID}
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CollectionIn: %v", err)
	}
	return in, buf
}
