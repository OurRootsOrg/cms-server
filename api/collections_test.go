package api_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"gocloud.dev/postgres"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/persist/dynamo"
	"github.com/stretchr/testify/assert"
)

func TestCollections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		db, err := postgres.Open(context.TODO(), databaseURL)
		if err != nil {
			log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
				err,
				databaseURL,
			)
		}
		p := persist.NewPostgresPersister(db)
		doCollectionsTests(t, p, p)
	}
	dynamoDBTableName := os.Getenv("DYNAMODB_TEST_TABLE_NAME")
	if dynamoDBTableName != "" {
		config := aws.Config{
			Region:      aws.String("us-east-1"),
			Endpoint:    aws.String("http://localhost:18000"),
			DisableSSL:  aws.Bool(true),
			Credentials: credentials.NewStaticCredentials("ACCESS_KEY", "SECRET", ""),
		}
		sess, err := session.NewSession(&config)
		assert.NoError(t, err)
		p, err := dynamo.NewPersister(sess, dynamoDBTableName)
		assert.NoError(t, err)
		doCollectionsTests(t, p, p)
	}

}
func doCollectionsTests(t *testing.T, catP model.CategoryPersister, colP model.CollectionPersister) {
	ap, err := api.NewAPI()
	assert.NoError(t, err)
	defer ap.Close()
	testApi := ap.
		CategoryPersister(catP).
		CollectionPersister(colP)
		// Add a test category for referential integrity
	testCategory := createTestCategory(t, catP)
	defer deleteTestCategory(t, catP, testCategory)

	// empty, err := testApi.GetCollections(context.TODO())
	// assert.NoError(t, err)
	// assert.Equal(t, 0, len(empty.Collections), "Expected empty slice, got %#v", empty)

	// Add a Collection
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{
			Name: "Test Collection",
		},
		Categories: []uint32{testCategory.ID},
	}
	created, err := testApi.AddCollection(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Categories, created.Categories)

	// Add with bad category reference
	in.Categories = []uint32{88}
	_, err = testApi.AddCollection(context.TODO(), in)
	assert.IsType(t, &api.Error{}, err)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrBadReference, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])

	// // GET /collections should now return the created Collection
	// ret, err := testApi.GetCollections(context.TODO())
	// assert.NoError(t, err)
	// assert.Equal(t, 0, len(empty.Collections), "Expected empty slice, got %#v", empty)
	// assert.Equal(t, 1, len(ret.Collections))
	// assert.Equal(t, *created, ret.Collections[0])

	// GET many collections should now return the created Collection
	colls, err := testApi.GetCollectionsByID(context.TODO(), []uint32{created.ID})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(colls))
	assert.Equal(t, *created, colls[0])

	// GET /collections/{id} should now return the created Collection
	ret2, err := testApi.GetCollection(context.TODO(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created, ret2)

	// Bad request - no category
	in.Categories = nil
	_, err = testApi.AddCollection(context.TODO(), in)
	assert.IsType(t, &api.Error{}, err)
	if assert.Len(t, err.(*api.Error).Errs(), 1, "err.(*api.Error).Errs(): %#v", err.(*api.Error).Errs()) {
		assert.Equal(t, err.(*api.Error).Errs()[0].Code, model.ErrRequired)
	}

	// Collection not found
	_, err = testApi.GetCollection(context.TODO(), created.ID+99)
	assert.Error(t, err)
	assert.IsType(t, &api.Error{}, err)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])

	// Update
	ret2.Name = "Updated"
	updated, err := testApi.UpdateCollection(context.TODO(), ret2.ID, *ret2)
	assert.NoError(t, err)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Categories, updated.Categories)
	assert.Equal(t, ret2.Name, updated.Name, "Expected Name to match")

	// Update non-existent
	_, err = testApi.UpdateCollection(context.TODO(), updated.ID+99, *updated)
	if assert.Error(t, err) {
		assert.IsType(t, &api.Error{}, err)
		assert.Len(t, err.(*api.Error).Errs(), 1)
		assert.Equal(t, model.ErrNotFound, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])
	}
	// Update with bad category
	updated.Categories = []uint32{999999}
	_, err = testApi.UpdateCollection(context.TODO(), updated.ID, *updated)
	if assert.Error(t, err) {
		assert.IsType(t, &api.Error{}, err)
		assert.Len(t, err.(*api.Error).Errs(), 1)
		assert.Equal(t, model.ErrBadReference, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])
	}
	// Update with bad LastUpdateTime
	updated.Categories = ret2.Categories
	updated.LastUpdateTime = time.Now().Add(-time.Minute)
	_, err = testApi.UpdateCollection(context.TODO(), updated.ID, *updated)
	if assert.Error(t, err) {
		assert.IsType(t, &api.Error{}, err)
		assert.Len(t, err.(*api.Error).Errs(), 1)
		assert.Equal(t, model.ErrConcurrentUpdate, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])
	}

	// DELETE
	err = testApi.DeleteCollection(context.TODO(), updated.ID)
	assert.NoError(t, err)
	_, err = testApi.GetCollection(context.TODO(), created.ID)
	assert.Error(t, err)
	assert.IsType(t, &api.Error{}, err)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])
}

func createTestCategory(t *testing.T, p model.CategoryPersister) *model.Category {
	in, err := model.NewCategoryIn("Test")
	assert.NoError(t, err)
	created, e := p.InsertCategory(context.TODO(), in)
	assert.NoError(t, e)
	return created
}

func deleteTestCategory(t *testing.T, p model.CategoryPersister, category *model.Category) {
	e := p.DeleteCategory(context.TODO(), category.ID)
	assert.NoError(t, e)
}
