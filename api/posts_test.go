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

func TestPosts(t *testing.T) {
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
		doPostsTests(t, p, p, p, p)
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
		doPostsTests(t, p, p, p, nil)
	}
}
func doPostsTests(t *testing.T,
	catP model.CategoryPersister,
	colP model.CollectionPersister,
	postP model.PostPersister,
	recordP model.RecordPersister,
) {
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		QueueConfig("recordswriter", "amqp://guest:guest@localhost:35672/").
		QueueConfig("publisher", "amqp://guest:guest@localhost:35672/").
		CollectionPersister(colP).
		PostPersister(postP).
		RecordPersister(recordP)

	// Add a test category and test collection for referential integrity
	testCategory := createTestCategory(t, catP)
	defer deleteTestCategory(t, catP, testCategory)
	testCollection := createTestCollection(t, colP, testCategory.ID)
	defer deleteTestCollection(t, colP, testCollection)

	empty, errors := testApi.GetPosts(context.TODO())
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(empty.Posts), "Expected empty slice, got %#v", empty)

	// Add a Post
	in := model.PostIn{
		PostBody: model.PostBody{
			Name: "Test Post",
		},
		Collection: testCollection.ID,
	}
	created, errors := testApi.AddPost(context.TODO(), in)
	assert.Nil(t, errors)
	defer deleteTestPost(t, postP, created)
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Collection, created.Collection)

	// Add with bad collection reference
	in.Collection = in.Collection + 88
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
	in.Collection = 0
	_, errors = testApi.AddPost(context.TODO(), in)
	if assert.Len(t, errors.Errs(), 1, "errors.Errs(): %#v", errors.Errs()) {
		assert.Equal(t, errors.Errs()[0].Code, model.ErrRequired)
	}

	// Post not found
	_, errors = testApi.GetPost(context.TODO(), created.ID+99)
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
	_, errors = testApi.UpdatePost(context.TODO(), updated.ID+99, *updated)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])

	// Update with bad collection
	updated.Collection = updated.Collection + 99
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
	_, errors = testApi.GetPost(context.TODO(), updated.ID)
	assert.NotNil(t, errors)
	assert.Len(t, errors.Errs(), 1)
	assert.Equal(t, model.ErrNotFound, errors.Errs()[0].Code, "errors.Errs()[0]: %#v", errors.Errs()[0])
}

func createTestCollection(t *testing.T, p model.CollectionPersister, categoryID uint32) *model.Collection {
	in := model.NewCollectionIn("Test", []uint32{categoryID})
	in.Fields = []model.CollectionField{
		{
			Header: "given",
		},
		{
			Header: "surname",
		},
	}
	in.Mappings = []model.CollectionMapping{
		{
			Header:  "given",
			DbField: "Given",
			IxRole:  "principal",
			IxField: "given",
		},
		{
			Header:  "surname",
			DbField: "Surname",
			IxRole:  "principal",
			IxField: "surname",
		},
	}
	created, e := p.InsertCollection(context.TODO(), in)
	assert.Nil(t, e)
	return created
}

func deleteTestCollection(t *testing.T, p model.CollectionPersister, collection *model.Collection) {
	e := p.DeleteCollection(context.TODO(), collection.ID)
	assert.Nil(t, e)
}
