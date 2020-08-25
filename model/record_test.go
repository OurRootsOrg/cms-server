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
	c := model.NewRecord(uint32(rand.Int31()), in)
	err = json.Unmarshal(js, &c)
	assert.NoError(t, err)
	// log.Printf("Record: %#v", c)
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Record JSON: %s", string(js))
	c.Post = 999
	js, err = json.Marshal(c)
	assert.NoError(t, err)
	// log.Printf("Record JSON: %s", string(js))
}

func TestRecordCitation(t *testing.T) {
	in := model.RecordIn{}
	in.Data = map[string]string{
		"Given name": "Fred",
		"Surname":    "Flintstone",
	}
	r := model.NewRecord(uint32(rand.Int31()), in)
	citationTemplate := `{{ Given name }} {{Surname}} {{ Missing Variable }} found in <i>My Favorite Collection</i>.`
	assert.Equal(t, "Fred Flintstone  found in <i>My Favorite Collection</i>.", r.GetCitation(citationTemplate))

	citationTemplate = `{{Invalid`
	assert.Equal(t, "template: citation:1: function \"Invalid\" not defined", r.GetCitation(citationTemplate))
}
