package model_test

import (
	"encoding/json"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestCategory(t *testing.T) {
	s, err := model.NewFieldDef("string", model.StringType)
	assert.NoError(t, err)
	ci, err := model.NewCategoryInput("Test Category", s)
	assert.NoError(t, err)
	js, err := json.Marshal(ci)
	assert.NoError(t, err)
	// log.Printf("Category JSON: %s", string(js))
	ci, err = model.NewCategoryInput("")
	assert.NoError(t, err)
	cat := model.NewCategory(ci)
	err = json.Unmarshal(js, &cat)
	assert.NoError(t, err)
	// log.Printf("Category: %#v", cat)

}
