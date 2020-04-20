package model_test

import (
	"testing"

	"github.com/jancona/ourroots/model"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	e := model.Error{Code: 200, Message: "Error message"}
	assert.NotNil(t, e)
}
