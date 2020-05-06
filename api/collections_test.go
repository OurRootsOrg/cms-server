package api_test

import (
	"context"
	"log"
	"os"
	"testing"

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
	p := persist.NewPostgresPersister("", db)
	testApi := api.NewAPI().
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
		Category: testCategory.CategoryRef,
	}
	created, errors := testApi.AddCollection(context.TODO(), in)
	assert.Nil(t, errors)
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Category, created.Category)

	// Add with bad category reference
	in.Category.ID = in.Category.ID + "88"
	_, errors = testApi.AddCollection(context.TODO(), in)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, api.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// GET /collections should now return the created Collection
	ret, errors := testApi.GetCollections(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Collections), "Expected empty slice, got %#v", empty)
	assert.Equal(t, 1, len(ret.Collections))
	assert.Equal(t, *created, ret.Collections[0])

	// GET /collections/{id} should now return the created Collection
	ret2, errors := testApi.GetCollection(context.TODO(), created.ID)
	assert.Nil(t, errors)
	assert.Equal(t, created, ret2)

	// Bad request - no category
	in.Category = model.CategoryRef{}
	_, errors = testApi.AddCollection(context.TODO(), in)
	assert.Len(t, errors.Errs(), 2, "errors.Errs(): %#v", errors.Errs())
	assert.Equal(t, errors.Errs()[0].Code, api.ErrRequired)
	assert.Equal(t, errors.Errs()[1].Code, api.ErrRequired)

	// Collection not found
	_, errors = testApi.GetCollection(context.TODO(), created.ID+"99")
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, api.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update
	ret2.Name = "Updated"
	updated, errors := testApi.UpdateCollection(context.TODO(), ret2.ID, ret2.CollectionIn)
	assert.Nil(t, errors)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Category, updated.Category)
	assert.Equal(t, ret2.Name, updated.Name, "Expected Name to match")

	// Update non-existant
	_, errors = testApi.UpdateCollection(context.TODO(), created.ID+"99", created.CollectionIn)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, api.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	created.Category.ID = created.Category.ID + "99"

	// Update with bad category
	_, errors = testApi.UpdateCollection(context.TODO(), created.ID, created.CollectionIn)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, api.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// DELETE
	errors = testApi.DeleteCollection(context.TODO(), created.ID)
	assert.Nil(t, errors)
}

func createTestCategory(p model.CategoryPersister) (*model.Category, error) {
	stringType, err := model.NewFieldDef("stringField", model.StringType, "string_field")
	if err != nil {
		return nil, err
	}
	in, err := model.NewCategoryIn("Test", stringType)
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
