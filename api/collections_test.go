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

func TestCollections(t *testing.T) {
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
	testApi := ap.
		CategoryPersister(p).
		CollectionPersister(p)

	// Add a test category for referential integrity
	testCategory, err := createTestCategory(p)
	assert.Nil(t, err, "Error creating test category")
	defer deleteTestCategory(p, testCategory)

	empty, errors := testApi.GetCollections(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Collections), "Expected empty slice, got %#v", empty)

	// Add a Collection
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{
			Name: "Test Collection",
		},
		Categories: []uint32{testCategory.ID},
	}
	created, errors := testApi.AddCollection(context.TODO(), in)
	assert.Nil(t, errors)
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Categories, created.Categories)

	// Add with bad category reference
	in.Categories = []uint32{88}
	_, errors = testApi.AddCollection(context.TODO(), in)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// GET /collections should now return the created Collection
	ret, errors := testApi.GetCollections(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Collections), "Expected empty slice, got %#v", empty)
	assert.Equal(t, 1, len(ret.Collections))
	assert.Equal(t, *created, ret.Collections[0])

	// GET many collections should now return the created Collection
	colls, errors := testApi.GetCollectionsByID(context.TODO(), []uint32{created.ID})
	assert.Nil(t, errors)
	assert.Equal(t, 1, len(colls))
	assert.Equal(t, *created, colls[0])

	// GET /collections/{id} should now return the created Collection
	ret2, errors := testApi.GetCollection(context.TODO(), created.ID)
	assert.Nil(t, errors)
	assert.Equal(t, created, ret2)

	// Bad request - no category
	in.Categories = nil
	_, errors = testApi.AddCollection(context.TODO(), in)
	if assert.Len(t, errors.Errs(), 1, "errors.Errs(): %#v", errors.Errs()) {
		assert.Equal(t, errors.Errs()[0].Code, model.ErrRequired)
	}

	// Collection not found
	_, errors = testApi.GetCollection(context.TODO(), created.ID+99)
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update
	ret2.Name = "Updated"
	updated, errors := testApi.UpdateCollection(context.TODO(), ret2.ID, *ret2)
	assert.Nil(t, errors)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Categories, updated.Categories)
	assert.Equal(t, ret2.Name, updated.Name, "Expected Name to match")

	// Update non-existant
	_, errors = testApi.UpdateCollection(context.TODO(), updated.ID+99, *updated)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad category
	updated.Categories = []uint32{99}
	_, errors = testApi.UpdateCollection(context.TODO(), updated.ID, *updated)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad LastUpdateTime
	updated.Categories = ret2.Categories
	updated.LastUpdateTime = time.Now().Add(-time.Minute)
	_, errors = testApi.UpdateCollection(context.TODO(), updated.ID, *updated)
	if assert.NotNil(t, errors) {
		assert.Len(t, errors.Errs(), 1)
		assert.Equal(t, model.ErrConcurrentUpdate, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])
	}

	// DELETE
	errors = testApi.DeleteCollection(context.TODO(), updated.ID)
	assert.Nil(t, errors)
	_, errors = testApi.GetCollection(context.TODO(), created.ID)
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])
}

func createTestCategory(p model.CategoryPersister) (*model.Category, error) {
	in, err := model.NewCategoryIn("Test")
	if err != nil {
		return nil, err
	}
	created, err := p.InsertCategory(context.TODO(), in)
	if err != nil {
		return nil, err
	}
	return &created, err
}

func deleteTestCategory(p model.CategoryPersister, category *model.Category) error {
	return p.DeleteCategory(context.TODO(), category.ID)
}
