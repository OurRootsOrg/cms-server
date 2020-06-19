package model_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestCollection(t *testing.T) {
	in := model.CollectionIn{}
	n := "Collection"
	in.Name = n
	js, err := json.Marshal(in)
	assert.NoError(t, err)
	// log.Printf("CollectionBody JSON: %s", string(js))
	in = model.CollectionIn{}
	c := model.NewCollection(uint32(rand.Int31()), in)
	err = json.Unmarshal(js, &c)
	assert.NoError(t, err)
	// log.Printf("Collection: %#v", c)
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Collection JSON: %s", string(js))
	c.Category = 999
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Collection JSON: %s", string(js))

}
