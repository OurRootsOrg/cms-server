package model_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	userIn := makeUserIn(t)
	js, err := json.Marshal(userIn)
	assert.NoError(t, err)
	log.Printf("UserBody JSON: %s", string(js))
	var user model.User
	err = json.Unmarshal(js, &user)
	assert.NoError(t, err)
	// log.Printf("User: %#v", cat)
	js, err = json.Marshal(user)
	assert.NoError(t, err)
}

func makeUserIn(t *testing.T) model.UserIn {
	in, err := model.NewUserIn("Test User", "user@example.com", false, "https://ourroots-jim.auth0.com/", "testsubject1")
	assert.NoError(t, err)
	return in
}
