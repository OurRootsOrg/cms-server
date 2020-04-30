package model_test

import (
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	e := model.NewError(model.ErrRequired, "some_field")
	assert.NotNil(t, e)
}
