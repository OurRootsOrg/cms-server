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

func TestGetRecordsForPost(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.RecordResult{}
	am.Result = &cr
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/records?post=1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty api.RecordResult
	err := json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty.Records), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Non-empty result
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	ci, _ := makeRecordIn(t)
	cr = api.RecordResult{
		Records: []model.Record{
			{
				ID:             1,
				RecordIn:       ci,
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.Result = &cr
	am.Errors = nil
	request, _ = http.NewRequest("GET", "/records?post=1", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret api.RecordResult
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret.Records))
	assert.Equal(t, cr.Records[0], ret.Records[0])

	// error result
	am.Result = (*api.RecordResult)(nil)
	am.Errors = api.NewErrors(http.StatusInternalServerError, assert.AnError)
	request, _ = http.NewRequest("GET", "/records?post=1", nil)
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
	assert.Equal(t, am.Errors.(*api.Errors).Errs(), errRet)
}

func makeRecordIn(t *testing.T) (model.RecordIn, *bytes.Buffer) {
	in := model.RecordIn{
		RecordBody: model.RecordBody{
			Data: map[string]string{
				"foo": "bar",
			},
		},
		Post: 1,
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding RecordIn: %v", err)
	}
	return in, buf
}
