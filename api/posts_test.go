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

func TestPosts(t *testing.T) {
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
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	testApi = testApi.
		CollectionPersister(p).
		PostPersister(p)

	// Add a test category and test collection for referential integrity
	testCategory, err := createTestCategory(p)
	assert.Nil(t, err, "Error creating test category")
	defer deleteTestCategory(p, testCategory)
	testCollection, err := createTestCollection(p, testCategory.ID)
	assert.Nil(t, err, "Error creating test collection")
	defer deleteTestCollection(p, testCollection)

	empty, errors := testApi.GetPosts(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Posts), "Expected empty slice, got %#v", empty)

	// Add a Post
	in := model.PostIn{
		PostBody: model.PostBody{
			Name:       "Test Post",
			RecordsKey: "key",
		},
		Collection: testCollection.ID,
	}
	created, errors := testApi.AddPost(context.TODO(), in)
	assert.Nil(t, errors)
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Collection, created.Collection)

	// Add with bad collection reference
	in.Collection = in.Collection + "88"
	_, errors = testApi.AddPost(context.TODO(), in)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// GET /posts should now return the created Post
	ret, errors := testApi.GetPosts(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Posts), "Expected empty slice, got %#v", empty)
	assert.Equal(t, 1, len(ret.Posts))
	assert.Equal(t, *created, ret.Posts[0])

	// GET /posts/{id} should now return the created Post
	ret2, errors := testApi.GetPost(context.TODO(), created.ID)
	assert.Nil(t, errors)
	assert.Equal(t, created, ret2)

	// Bad request - no collection
	in.Collection = ""
	_, errors = testApi.AddPost(context.TODO(), in)
	if assert.Len(t, errors.Errs(), 1, "errors.Errs(): %#v", errors.Errs()) {
		assert.Equal(t, errors.Errs()[0].Code, model.ErrRequired)
	}

	// Post not found
	_, errors = testApi.GetPost(context.TODO(), created.ID+"99")
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update
	ret2.Name = "Updated"
	updated, errors := testApi.UpdatePost(context.TODO(), ret2.ID, *ret2)
	assert.Nil(t, errors)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Collection, updated.Collection)
	assert.Equal(t, ret2.Name, updated.Name, "Expected Name to match")

	// Update non-existant
	_, errors = testApi.UpdatePost(context.TODO(), updated.ID+"99", *updated)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad collection
	updated.Collection = updated.Collection + "99"
	_, errors = testApi.UpdatePost(context.TODO(), updated.ID, *updated)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad LastUpdateTime
	updated.Collection = ret2.Collection
	updated.LastUpdateTime = time.Now().Add(-time.Minute)
	_, errors = testApi.UpdatePost(context.TODO(), updated.ID, *updated)
	if assert.NotNil(t, errors) {
		assert.Len(t, errors.Errs(), 1)
		assert.Equal(t, model.ErrConcurrentUpdate, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])
	}

	// DELETE
	errors = testApi.DeletePost(context.TODO(), updated.ID)
	assert.Nil(t, errors)
}

func createTestCollection(p model.CollectionPersister, categoryID string) (*model.Collection, error) {
	in := model.NewCollectionIn("Test", categoryID)
	created, err := p.InsertCollection(context.TODO(), in)
	if err != nil {
		return nil, err
	}
	return &created, err
}

func deleteTestCollection(p model.CollectionPersister, collection *model.Collection) error {
	return p.DeleteCollection(context.TODO(), collection.ID)
}
