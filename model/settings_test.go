package model_test

import (
	"encoding/json"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestSettings(t *testing.T) {
	in := model.SettingsIn{}
	in.PostFields = []model.SettingsPostField{
		{
			Name: "One",
			Type: "string",
		},
		{
			Name: "Two",
			Type: "number",
		},
	}
	js, err := json.Marshal(in)
	assert.NoError(t, err)
	in = model.SettingsIn{}
	c := model.NewSettings(in)
	err = json.Unmarshal(js, &c)
	assert.NoError(t, err)
	js, err = json.Marshal(c)
	assert.NoError(t, err)
}
