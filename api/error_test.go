package api_test

import (
	"testing"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	e := api.NewError(api.ErrRequired, "some_field")
	assert.NotNil(t, e)
}
