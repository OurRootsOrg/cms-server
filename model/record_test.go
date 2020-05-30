package model_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestRecord(t *testing.T) {
	in := model.RecordIn{}
	in.Data = map[string]string{
		"foo": "bar",
	}
	js, err := json.Marshal(in)
	assert.NoError(t, err)
	// log.Printf("RecordBody JSON: %s", string(js))
	in = model.RecordIn{}
	c := model.NewRecord(int32(rand.Int31()), in)
	err = json.Unmarshal(js, &c)
	assert.NoError(t, err)
	// log.Printf("Record: %#v", c)
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Record JSON: %s", string(js))
	c.Post = "/posts/999"
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Record JSON: %s", string(js))
}
