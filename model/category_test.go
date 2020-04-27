package model_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestCategory(t *testing.T) {
	cb := makeCategoryIn(t)
	js, err := json.Marshal(cb)
	assert.NoError(t, err)
	log.Printf("CategoryBody JSON: %s", string(js))
	var cat model.Category
	err = json.Unmarshal(js, &cat)
	assert.NoError(t, err)
	// log.Printf("Category: %#v", cat)
	js, err = json.Marshal(cat)
	assert.NoError(t, err)
	// log.Printf("Category JSON: %s", string(js))
	intType, err := model.NewFieldDef("intField", model.IntType, "int_field")
	assert.NoError(t, err)
	_, err = model.NewCategoryIn("Test Category", intType, intType)
	assert.Error(t, err)
}

func makeCategoryIn(t *testing.T) model.CategoryIn {
	intType, err := model.NewFieldDef("intField", model.IntType, "int_field")
	assert.NoError(t, err)
	stringType, err := model.NewFieldDef("stringField", model.StringType, "string_field")
	assert.NoError(t, err)
	imageType, err := model.NewFieldDef("imageField", model.ImageType, "image_field")
	assert.NoError(t, err)
	locationType, err := model.NewFieldDef("locationField", model.LocationType, "location_field")
	assert.NoError(t, err)
	timeType, err := model.NewFieldDef("timeField", model.TimeType, "time_field")
	assert.NoError(t, err)
	in, err := model.NewCategoryIn("Test Category", intType, stringType, imageType, locationType, timeType)
	assert.NoError(t, err)
	return in
}
