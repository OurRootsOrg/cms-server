package model_test

import (
	"encoding/json"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestCollection(t *testing.T) {
	ci := model.CollectionInput{
		Name: "Collection",
	}
	js, err := json.Marshal(ci)
	assert.NoError(t, err)
	// log.Printf("CollectionInput JSON: %s", string(js))
	ci = model.CollectionInput{}
	assert.NoError(t, err)
	c := model.NewCollection(ci)
	err = json.Unmarshal(js, &c)
	assert.NoError(t, err)
	// log.Printf("Collection: %#v", c)
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Collection JSON: %s", string(js))
	cati, err := model.NewCategoryInput("")
	assert.NoError(t, err)
	cat := model.NewCategory(cati)
	c.Category = &cat
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Collection JSON: %s", string(js))

}
