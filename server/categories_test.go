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
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.CategoryResult{}
	am.result = &cr
	am.errors = nil

	request, _ := http.NewRequest("GET", "/categories", nil)
	response := httptest.NewRecorder()
	request.Header.Add("Authorization", "Bearer XYZ")
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
				ID: "/categories/1",
				CategoryBody: model.CategoryBody{
					Name: "Test name",
				},
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.result = &cr
	am.errors = nil
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
	am.result = (*api.CategoryResult)(nil)
	am.errors = model.NewErrors(http.StatusInternalServerError, assert.AnError)
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
	assert.Equal(t, am.errors.Errs(), errRet)
}
func TestGetCategory(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	r := app.NewRouter()

	category := &model.Category{
		ID: "/categories/1",
		CategoryBody: model.CategoryBody{
			Name: "Name",
		},
	}
	am.result = category
	am.errors = nil
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
	am.result = category
	am.errors = model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, "/categories/1"))

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
	assert.Equal(t, am.errors.Errs(), errRet)
}

func TestPostCategory(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	r := app.NewRouter()

	in, buf := makeCategoryIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	am.result = &model.Category{
		ID:             "/categories/1",
		CategoryBody:   in.CategoryBody,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.errors = nil

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
	assert.Equal(t, in.FieldDefs, created.FieldDefs, "Expected FieldDefs to match")
	assert.NotEmpty(t, created.ID)
}

func TestPutCategory(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	r := app.NewRouter()

	in, buf := makeCategoryIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	cat := model.Category{
		ID:             "/categories/1",
		CategoryBody:   in.CategoryBody,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.result = &cat
	am.errors = nil

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
	assert.Equal(t, in.FieldDefs, created.FieldDefs, "Expected FieldDefs to match")
	assert.NotEmpty(t, created.ID)
}

func TestDeleteCategory(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	r := app.NewRouter()

	am.result = nil
	am.errors = nil

	request, _ := http.NewRequest("DELETE", "/categories/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code, "Response: %s", string(response.Body.Bytes()))
}

func makeCategoryIn(t *testing.T) (model.CategoryIn, *bytes.Buffer) {
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
	return in, buf
}
