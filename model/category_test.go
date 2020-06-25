package model_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
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
}

func makeCategoryIn(t *testing.T) model.CategoryIn {
	in, err := model.NewCategoryIn("Test Category")
	assert.NoError(t, err)
	return in
}
