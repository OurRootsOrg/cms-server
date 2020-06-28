package api_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
)

func TestCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	db, err := postgres.Open(context.TODO(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
			err,
			os.Getenv("DATABASE_URL"),
		)
	}
	p := persist.NewPostgresPersister(db)
	ap, err := api.NewAPI()
	assert.NoError(t, err)
	defer ap.Close()
	testApi := ap.CategoryPersister(p)

	empty, errors := testApi.GetCategories(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Categories), "Expected empty slice, got %#v", empty)

	// Add a Category
	in := model.CategoryIn{
		CategoryBody: model.CategoryBody{
			Name: "Test Category",
		},
	}
	created, errors := testApi.AddCategory(context.TODO(), in)
	assert.Nil(t, errors)
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)

	// GET /collections should now return the created Category
	ret, errors := testApi.GetCategories(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Categories), "Expected empty slice, got %#v", empty)
	assert.Equal(t, 1, len(ret.Categories))
	assert.Equal(t, *created, ret.Categories[0])

	// GET /collections/{id} should now return the created Category
	ret2, errors := testApi.GetCategory(context.TODO(), created.ID)
	assert.Nil(t, errors)
	assert.Equal(t, created, ret2)

	// Category not found
	_, errors = testApi.GetCategory(context.TODO(), created.ID+99)
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update
	ret2.Name = "Updated"
	updated, errors := testApi.UpdateCategory(context.TODO(), ret2.ID, *ret2)
	assert.Nil(t, errors)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Name, updated.Name, "Expected Name to match")

	// Update non-existent
	_, errors = testApi.UpdateCategory(context.TODO(), created.ID+99, *created)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad LastUpdateTime
	updated.LastUpdateTime = time.Now().Add(-time.Minute)
	updated, errors = testApi.UpdateCategory(context.TODO(), updated.ID, *updated)
	if assert.NotNil(t, errors) {
		assert.Len(t, errors.Errs(), 1)
		assert.Equal(t, model.ErrConcurrentUpdate, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])
	}

	// DELETE
	errors = testApi.DeleteCategory(context.TODO(), created.ID)
	assert.Nil(t, errors)
	_, errors = testApi.GetCategory(context.TODO(), created.ID)
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])
}
