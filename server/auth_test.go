package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockVerifier struct {
	mock.Mock
}

func (m *mockVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	rets := m.Called(ctx, rawIDToken)
	return rets[0].(*oidc.IDToken), rets.Error(1)
}

func TestAuth(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	m := mockVerifier{}
	app.oidcVerifier = &m
	var requestContext context.Context

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.Handle(app.baseURL.Path+"/health",
		app.verifyToken(http.HandlerFunc(
			func(rw http.ResponseWriter, r *http.Request) {
				requestContext = r.Context()
				log.Print("[DEBUG] Called")
				rw.WriteHeader(http.StatusOK)
			}))).Methods("GET")

	// No Auth header
	request, _ := http.NewRequest("GET", "/health", nil)
	response := httptest.NewRecorder()
	// request.Header.Add("Authorization", "Bearer XYZ")
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Bad Auth header
	request, _ = http.NewRequest("GET", "/health", nil)
	response = httptest.NewRecorder()
	request.Header.Add("Authorization", "Abc")
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	request, _ = http.NewRequest("GET", "/health", nil)
	response = httptest.NewRecorder()
	request.Header.Add("Authorization", "Abc Def")
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Bad token
	m.On("Verify", mock.Anything, "Abc").Once().Return((*oidc.IDToken)(nil), errors.New("Bad token format"))
	request, _ = http.NewRequest("GET", "/health", nil)
	response = httptest.NewRecorder()
	request.Header.Add("Authorization", "Bearer Abc")
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	m.AssertExpectations(t)

	// "Good" token
	expectedUser := model.User{}
	am.Result = &expectedUser
	am.Errors = nil

	parsedToken := oidc.IDToken{}
	m.On("Verify", mock.Anything, "Abc").Once().Return(&parsedToken, nil)
	request, _ = http.NewRequest("GET", "/health", nil)
	response = httptest.NewRecorder()
	request.Header.Add("Authorization", "Bearer Abc")
	r.ServeHTTP(response, request)
	if assert.Equal(t, http.StatusOK, response.Code) {
		user := requestContext.Value(api.UserProperty)
		actualUser := user.(*model.User)
		assert.Equal(t, expectedUser, *actualUser)
	}
	m.AssertExpectations(t)
}
