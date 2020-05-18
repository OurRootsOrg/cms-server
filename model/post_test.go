package model_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	in := model.PostIn{}
	n := "Post"
	in.Name = n
	js, err := json.Marshal(in)
	assert.NoError(t, err)
	// log.Printf("PostBody JSON: %s", string(js))
	in = model.PostIn{}
	c := model.NewPost(int32(rand.Int31()), in)
	err = json.Unmarshal(js, &c)
	assert.NoError(t, err)
	// log.Printf("Post: %#v", c)
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Post JSON: %s", string(js))
	c.Collection = "/collections/999"
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Post JSON: %s", string(js))
}
