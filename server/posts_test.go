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

func TestGetAllPosts(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.PostResult{}
	am.result = &cr
	am.errors = nil

	request, _ := http.NewRequest("GET", "/posts", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty api.PostResult
	err := json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty.Posts), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Non-empty result
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	ci, _ := makePostIn(t)
	cr = api.PostResult{
		Posts: []model.Post{
			{
				ID:             "/posts/1",
				PostIn:         ci,
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.result = &cr
	am.errors = nil
	request, _ = http.NewRequest("GET", "/posts", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret api.PostResult
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret.Posts))
	assert.Equal(t, cr.Posts[0], ret.Posts[0])

	// error result
	am.result = (*api.PostResult)(nil)
	am.errors = model.NewErrors(http.StatusInternalServerError, assert.AnError)
	request, _ = http.NewRequest("GET", "/posts", nil)
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
func TestGetPost(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	ci, _ := makePostIn(t)
	post := &model.Post{
		ID:     "/posts/1",
		PostIn: ci,
	}
	am.result = post
	am.errors = nil
	var ret model.Post

	request, _ := http.NewRequest("GET", "/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	err := json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, *post, ret)

	post = nil
	am.result = post
	am.errors = model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, "/posts/1"))

	request, _ = http.NewRequest("GET", "/posts/1", nil)
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

func TestPostPost(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makePostIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	am.result = &model.Post{
		ID:             "/posts/1",
		PostIn:         in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.errors = nil

	request, _ := http.NewRequest("POST", "/posts", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Post
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.Equal(t, in, created.PostIn)
	assert.Equal(t, now, created.InsertTime)
	assert.Equal(t, now, created.LastUpdateTime)
	assert.NotEmpty(t, created.ID)
}

func TestPutPost(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makePostIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	post := model.Post{
		ID:             "/posts/1",
		PostIn:         in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.result = &post
	am.errors = nil

	request, _ := http.NewRequest("PUT", "/posts/1", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Post
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.Equal(t, in, created.PostIn)
	assert.Equal(t, now, created.InsertTime)
	assert.Equal(t, now, created.LastUpdateTime)
	assert.NotEmpty(t, created.ID)
}

func TestDeletePost(t *testing.T) {
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	am.result = nil
	am.errors = nil

	request, _ := http.NewRequest("DELETE", "/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code, "Response: %s", string(response.Body.Bytes()))
}

func makePostIn(t *testing.T) (model.PostIn, *bytes.Buffer) {
	in := model.PostIn{
		PostBody: model.PostBody{
			Name:          "First",
			RecordsKey:    "key",
			RecordsStatus: api.PostDraft,
		},
		Collection: "/collections/1",
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding PostIn: %v", err)
	}
	return in, buf
}
