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

func TestCategories(t *testing.T) {
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
		doCategoriesTests(t, p)
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
		doCategoriesTests(t, p)
	}

}
func doCategoriesTests(t *testing.T, p model.CategoryPersister) {
	ap, err := api.NewAPI()
	assert.NoError(t, err)
	defer ap.Close()
	testApi := ap.CategoryPersister(p)
	empty, err := testApi.GetCategories(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(empty.Categories), "Expected empty slice, got %#v", empty)

	// Add a Category
	in := model.CategoryIn{
		CategoryBody: model.CategoryBody{
			Name: "Test Category",
		},
	}
	created, err := testApi.AddCategory(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.NotEmpty(t, created.ID)

	// GET /collections should now return the created Category
	ret, err := testApi.GetCategories(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(empty.Categories), "Expected empty slice, got %#v", empty)
	assert.Equal(t, 1, len(ret.Categories))
	assert.Equal(t, *created, ret.Categories[0])

	// GET /collections/{id} should now return the created Category
	ret2, err := testApi.GetCategory(context.TODO(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created, ret2)

	// Category not found
	_, err = testApi.GetCategory(context.TODO(), created.ID+99)
	assert.Error(t, err)
	assert.IsType(t, &model.Errors{}, err)
	assert.Len(t, err.(*model.Errors).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*model.Errors).Errs()[0].Code, "err.(*model.Errors).Errs()[0]: %#v", err.(*model.Errors).Errs()[0])

	// Update
	ret2.Name = "Updated"
	updated, err := testApi.UpdateCategory(context.TODO(), ret2.ID, *ret2)
	assert.NoError(t, err)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Name, updated.Name, "Expected Name to match")

	// Update non-existent
	_, err = testApi.UpdateCategory(context.TODO(), created.ID+99, *created)
	assert.IsType(t, &model.Errors{}, err)
	assert.Len(t, err.(*model.Errors).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*model.Errors).Errs()[0].Code, "err.(*model.Errors).Errs()[0]: %#v", err.(*model.Errors).Errs()[0])

	// Update with bad LastUpdateTime
	updated.LastUpdateTime = time.Now().Add(-time.Minute)
	updated, err = testApi.UpdateCategory(context.TODO(), updated.ID, *updated)
	if assert.Error(t, err) {
		assert.IsType(t, &model.Errors{}, err)
		assert.Len(t, err.(*model.Errors).Errs(), 1)
		assert.Equal(t, model.ErrConcurrentUpdate, err.(*model.Errors).Errs()[0].Code, "err.(*model.Errors).Errs()[0]: %#v", err.(*model.Errors).Errs()[0])
	}

	// DELETE
	err = testApi.DeleteCategory(context.TODO(), created.ID)
	assert.NoError(t, err)
	_, err = testApi.GetCategory(context.TODO(), created.ID)
	assert.Error(t, err)
	assert.IsType(t, &model.Errors{}, err)
	assert.Len(t, err.(*model.Errors).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*model.Errors).Errs()[0].Code, "err.(*model.Errors).Errs()[0]: %#v", err.(*model.Errors).Errs()[0])
}
