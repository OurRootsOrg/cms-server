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

func TestGetAllCategories(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.CategoryResult{}
	am.Result = &cr
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/categories", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty api.CategoryResult
	err := json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty.Categories), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Non-empty result
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	cr = api.CategoryResult{
		Categories: []model.Category{
			{
				ID: 1,
				CategoryBody: model.CategoryBody{
					Name: "Test name",
				},
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.Result = &cr
	am.Errors = nil
	request, _ = http.NewRequest("GET", "/categories", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret api.CategoryResult
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret.Categories))
	assert.Equal(t, cr.Categories[0], ret.Categories[0])

	// error result
	am.Result = (*api.CategoryResult)(nil)
	am.Errors = api.NewError(assert.AnError)
	request, _ = http.NewRequest("GET", "/categories", nil)
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
func TestGetCategory(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	category := &model.Category{
		ID: 1,
		CategoryBody: model.CategoryBody{
			Name: "Name",
		},
	}
	am.Result = category
	am.Errors = nil
	var ret model.Category

	request, _ := http.NewRequest("GET", "/categories/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	err := json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, *category, ret)

	category = nil
	am.Result = category
	am.Errors = api.NewError(model.NewError(model.ErrNotFound, "1"))

	request, _ = http.NewRequest("GET", "/categories/1", nil)
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

func TestPostCategory(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makeCategoryIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	am.Result = &model.Category{
		ID:             1,
		CategoryBody:   in.CategoryBody,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.Errors = nil

	request, _ := http.NewRequest("POST", "/categories", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Category
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
}

func TestPutCategory(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makeCategoryIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	cat := model.Category{
		ID:             1,
		CategoryBody:   in.CategoryBody,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.Result = &cat
	am.Errors = nil

	request, _ := http.NewRequest("PUT", "/categories/1", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Category
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
}

func TestDeleteCategory(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	am.Result = nil
	am.Errors = nil

	request, _ := http.NewRequest("DELETE", "/categories/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code, "Response: %s", string(response.Body.Bytes()))
}

func makeCategoryIn(t *testing.T) (model.CategoryIn, *bytes.Buffer) {
	in, err := model.NewCategoryIn("First")
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding CategoryIn: %v", err)
	}
	return in, buf
}
