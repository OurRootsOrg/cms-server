package model_test

import (
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	e := model.NewError(model.ErrRequired, "some_field")
	assert.NotNil(t, e)

	assert.True(t, model.ErrRequired.Matches(e))

	var err error
	err = model.NewError(model.ErrConflict)
	assert.False(t, model.ErrRequired.Matches(err))
	assert.True(t, model.ErrConflict.Matches(err))
}
