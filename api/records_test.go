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

func TestRecords(t *testing.T) {
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
	defer testApi.Close()
	testApi = testApi.
		CategoryPersister(p).
		CollectionPersister(p).
		PostPersister(p).
		RecordPersister(p)

	// Add a test category and test collection and test post for referential integrity
	testCategory, err := createTestCategory(p)
	assert.Nil(t, err, "Error creating test category")
	defer deleteTestCategory(p, testCategory)
	testCollection, err := createTestCollection(p, testCategory.ID)
	assert.Nil(t, err, "Error creating test collection")
	defer deleteTestCollection(p, testCollection)
	testPost, err := createTestPost(p, testCollection.ID)
	assert.Nil(t, err, "Error creating test post")
	defer deleteTestPost(p, testPost)

	empty, errors := testApi.GetRecordsForPost(context.TODO(), testPost.ID)
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Records), "Expected empty slice, got %#v", empty)

	// Add a Record
	in := model.RecordIn{
		RecordBody: model.RecordBody{
			Data: map[string]string{"foo": "bar"},
		},
		Post: testPost.ID,
	}
	created, errors := testApi.AddRecord(context.TODO(), in)
	assert.Nil(t, errors)
	assert.Equal(t, in.Data, created.Data, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Post, created.Post)

	// Add with bad post reference
	in.Post = in.Post + "88"
	_, errors = testApi.AddRecord(context.TODO(), in)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// GET /records should now return the created Record
	ret, errors := testApi.GetRecordsForPost(context.TODO(), testPost.ID)
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Records), "Expected empty slice, got %#v", empty)
	assert.Equal(t, 1, len(ret.Records))
	assert.Equal(t, *created, ret.Records[0])

	// GET /records/{id} should now return the created Record
	ret2, errors := testApi.GetRecord(context.TODO(), created.ID)
	assert.Nil(t, errors)
	assert.Equal(t, created, ret2)

	// Bad request - no post
	in.Post = ""
	_, errors = testApi.AddRecord(context.TODO(), in)
	if assert.Len(t, errors.Errs(), 1, "errors.Errs(): %#v", errors.Errs()) {
		assert.Equal(t, errors.Errs()[0].Code, model.ErrRequired)
	}

	// Record not found
	_, errors = testApi.GetRecord(context.TODO(), created.ID+"99")
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update
	ret2.Data = map[string]string{"foo": "baz"}
	updated, errors := testApi.UpdateRecord(context.TODO(), ret2.ID, *ret2)
	assert.Nil(t, errors)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Post, updated.Post)
	assert.Equal(t, ret2.Data, updated.Data, "Expected Name to match")

	// Update non-existant
	_, errors = testApi.UpdateRecord(context.TODO(), updated.ID+"99", *updated)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad post
	updated.Post = updated.Post + "99"
	_, errors = testApi.UpdateRecord(context.TODO(), updated.ID, *updated)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrBadReference, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad LastUpdateTime
	updated.Post = ret2.Post
	updated.LastUpdateTime = time.Now().Add(-time.Minute)
	_, errors = testApi.UpdateRecord(context.TODO(), updated.ID, *updated)
	if assert.NotNil(t, errors) {
		assert.Len(t, errors.Errs(), 1)
		assert.Equal(t, model.ErrConcurrentUpdate, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])
	}

	// DELETE
	errors = testApi.DeleteRecord(context.TODO(), updated.ID)
	assert.Nil(t, errors)
}

func createTestPost(p model.PostPersister, collectionID string) (*model.Post, error) {
	in := model.NewPostIn("Test", collectionID, "")
	created, err := p.InsertPost(context.TODO(), in)
	if err != nil {
		return nil, err
	}
	return &created, err
}

func deleteTestPost(p model.PostPersister, post *model.Post) error {
	return p.DeletePost(context.TODO(), post.ID)
}
