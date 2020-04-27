package model_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestCollection(t *testing.T) {
	in := model.CollectionIn{}
	in.Name = "Collection"
	js, err := json.Marshal(in)
	assert.NoError(t, err)
	// log.Printf("CollectionBody JSON: %s", string(js))
	in = model.CollectionIn{}
	assert.NoError(t, err)
	c := model.NewCollection(int32(rand.Int31()), in)
	err = json.Unmarshal(js, &c)
	assert.NoError(t, err)
	// log.Printf("Collection: %#v", c)
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Collection JSON: %s", string(js))
	cr := model.NewCategoryRef(999)
	c.Category = cr
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Collection JSON: %s", string(js))

}
