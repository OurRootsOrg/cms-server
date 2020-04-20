package model_test

import (
	"encoding/json"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestFieldDef(t *testing.T) {
	intType, err := model.NewFieldDef("intField", model.IntType)
	assert.NoError(t, err)
	assert.Equal(t, "intField", intType.Name())
	assert.Equal(t, model.IntType, intType.Type())
	stringType, err := model.NewFieldDef("stringField", model.StringType)
	assert.NoError(t, err)
	assert.Equal(t, "stringField", stringType.Name())
	assert.Equal(t, model.StringType, stringType.Type())
	imageType, err := model.NewFieldDef("imageField", model.ImageType)
	assert.NoError(t, err)
	assert.Equal(t, "imageField", imageType.Name())
	assert.Equal(t, model.ImageType, imageType.Type())
	locationType, err := model.NewFieldDef("locationField", model.LocationType)
	assert.NoError(t, err)
	assert.Equal(t, "locationField", locationType.Name())
	assert.Equal(t, model.LocationType, locationType.Type())
	timeType, err := model.NewFieldDef("timeField", model.TimeType)
	assert.NoError(t, err)
	assert.Equal(t, "timeField", timeType.Name())
	assert.Equal(t, model.TimeType, timeType.Type())
}
func TestFieldDefSet(t *testing.T) {
	fds := model.NewFieldDefSet()
	intType, err := model.NewFieldDef("intField", model.IntType)
	assert.NoError(t, err)
	assert.Equal(t, true, fds.Add(intType))
	assert.True(t, fds.Contains(intType))

	stringType, err := model.NewFieldDef("stringField", model.StringType)
	assert.NoError(t, err)
	assert.Equal(t, true, fds.Add(stringType))
	assert.True(t, fds.Contains(stringType))

	imageType, err := model.NewFieldDef("imageField", model.ImageType)
	assert.NoError(t, err)
	assert.Equal(t, true, fds.Add(imageType))
	assert.True(t, fds.Contains(imageType))

	locationType, err := model.NewFieldDef("locationField", model.LocationType)
	assert.NoError(t, err)
	assert.Equal(t, true, fds.Add(locationType))
	assert.True(t, fds.Contains(locationType))

	timeType, err := model.NewFieldDef("timeField", model.TimeType)
	assert.NoError(t, err)
	assert.Equal(t, true, fds.Add(timeType))
	assert.True(t, fds.Contains(timeType))

	js, err := json.Marshal(fds)
	assert.NoError(t, err)

	fds2 := model.NewFieldDefSet()
	err = json.Unmarshal([]byte(js), &fds2)
	assert.NoError(t, err)
	assert.Equal(t, fds, fds2)

	assert.Equal(t, false, fds.Add(timeType))
	assert.True(t, fds.Contains(timeType))

	badJson := `{"imageField":"Image","imageField":"String"}`
	err = json.Unmarshal([]byte(badJson), &fds2)
	assert.Error(t, err)

	badJson = `{"imageField":"BadImage"}`
	err = json.Unmarshal([]byte(badJson), &fds2)
	assert.Error(t, err)

	badJson = `{"name":"Test Category","field_defs":{"imageField":Image"}}`
	err = json.Unmarshal([]byte(badJson), &fds2)
	assert.Error(t, err)

	_, err = model.NewFieldDef("badType", "badType")
	assert.Error(t, err)

}
