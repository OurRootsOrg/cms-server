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
	am := &apiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.RecordResult{}
	am.result = &cr
	am.errors = nil

	request, _ := http.NewRequest("GET", "/records?post=/posts/1", nil)
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
				ID:             "/records/1",
				RecordIn:       ci,
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.result = &cr
	am.errors = nil
	request, _ = http.NewRequest("GET", "/records?post=/posts/1", nil)
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
	am.result = (*api.RecordResult)(nil)
	am.errors = model.NewErrors(http.StatusInternalServerError, assert.AnError)
	request, _ = http.NewRequest("GET", "/records?post=/posts/1", nil)
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

func makeRecordIn(t *testing.T) (model.RecordIn, *bytes.Buffer) {
	in := model.RecordIn{
		RecordBody: model.RecordBody{
			Data: map[string]string{
				"foo": "bar",
			},
		},
		Post: "/posts/1",
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding RecordIn: %v", err)
	}
	return in, buf
}
