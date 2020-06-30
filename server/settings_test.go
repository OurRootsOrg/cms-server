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

func TestGetSettings(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	ci, _ := makeSettingsIn(t)
	settings := &model.Settings{
		SettingsIn: ci,
	}
	am.Result = settings
	am.Errors = nil
	var ret model.Settings

	request, _ := http.NewRequest("GET", "/settings", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	err := json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, *settings, ret)

	settings = nil
	am.Result = settings
	am.Errors = model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, "/settings"))

	request, _ = http.NewRequest("GET", "/settings", nil)
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
	assert.Equal(t, am.Errors.Errs(), errRet)
}

func TestPutSettings(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makeSettingsIn(t)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	settings := model.Settings{
		SettingsIn:     in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.Result = &settings
	am.Errors = nil

	request, _ := http.NewRequest("PUT", "/settings", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Settings
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in, created.SettingsIn)
	assert.Equal(t, now, created.InsertTime)
	assert.Equal(t, now, created.LastUpdateTime)
}

func makeSettingsIn(t *testing.T) (model.SettingsIn, *bytes.Buffer) {
	in := model.SettingsIn{
		SettingsBody: model.SettingsBody{
			PostMetadata: []model.SettingsPostMetadata{
				{
					Name: "One",
					Type: "string",
				},
			},
		},
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding SettingsIn: %v", err)
	}
	return in, buf
}
