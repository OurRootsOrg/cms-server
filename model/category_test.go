package model_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestCategory(t *testing.T) {
	intType, err := model.NewFieldDef("intField", model.IntType)
	assert.NoError(t, err)
	stringType, err := model.NewFieldDef("stringField", model.StringType)
	assert.NoError(t, err)
	imageType, err := model.NewFieldDef("imageField", model.ImageType)
	assert.NoError(t, err)
	locationType, err := model.NewFieldDef("locationField", model.LocationType)
	assert.NoError(t, err)
	timeType, err := model.NewFieldDef("timeField", model.TimeType)
	assert.NoError(t, err)
	ci, err := model.NewCategoryInput("Test Category", intType, stringType, imageType, locationType, timeType)
	assert.NoError(t, err)
	js, err := json.Marshal(ci)
	assert.NoError(t, err)
	log.Printf("CategoryInput JSON: %s", string(js))
	var cat model.Category
	err = json.Unmarshal(js, &cat)
	assert.NoError(t, err)
	// log.Printf("Category: %#v", cat)
	js, err = json.Marshal(cat)
	assert.NoError(t, err)
	// log.Printf("Category JSON: %s", string(js))
	_, err = model.NewCategoryInput("Test Category", intType, intType)
	assert.Error(t, err)
}
