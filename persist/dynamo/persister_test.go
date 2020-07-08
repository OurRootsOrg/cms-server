package dynamo_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/persist/dynamo"
	"github.com/stretchr/testify/assert"
)

func setupTestCase(t *testing.T) (dynamo.Persister, func(t *testing.T)) {
	t.Log("setup test case")
	table := os.Getenv("DYNAMODB_TEST_TABLE_NAME")
	if table == "" {
		t.Log("No DYNAMODB_TEST_TABLE_NAME specified, skipping tests")
		os.Exit(1)
	}
	config := aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("http://localhost:18000"),
		DisableSSL:  aws.Bool(true),
		Credentials: credentials.NewStaticCredentials("ACCESS_KEY", "SECRET", ""),
	}
	sess, err := session.NewSession(&config)
	assert.NoError(t, err)
	p, err := dynamo.NewPersister(sess, table)
	assert.NoError(t, err)
	return p, func(t *testing.T) {
		t.Log("teardown test case")
		colls, err := p.SelectCollections(context.TODO())
		assert.NoError(t, err)

		for _, c := range colls {
			err = p.DeleteCollection(context.TODO(), c.ID)
			assert.NoError(t, err)
		}

		cats, err := p.SelectCategories(context.TODO())
		assert.NoError(t, err)

		for _, c := range cats {
			err = p.DeleteCategory(context.TODO(), c.ID)
			assert.NoError(t, err)
		}
	}
}

func TestCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)

	_, err := p.SelectOneCategory(context.TODO(), 1)
	assert.Error(t, err)
	assert.IsType(t, model.Error{}, err)
	assert.Equal(t, model.ErrNotFound, err.(model.Error).Code)
	assert.Equal(t, "1", err.(model.Error).Params[0])

	ci, err := model.NewCategoryIn("Test Category")
	assert.NoError(t, err)

	cat, err := p.InsertCategory(context.TODO(), ci)
	assert.NoError(t, err)
	assert.Equal(t, ci.Name, cat.Name)

	c, err := p.SelectOneCategory(context.TODO(), cat.ID)
	assert.NoError(t, err)
	assert.Equal(t, cat, c)

	// Add another
	ci, err = model.NewCategoryIn("Test Category 2")
	assert.NoError(t, err)

	cat2, err := p.InsertCategory(context.TODO(), ci)
	assert.NoError(t, err)
	assert.Equal(t, ci.Name, cat2.Name)

	// Get both in one call
	cats, err := p.SelectCategoriesByID(context.TODO(), []uint32{cat.ID, cat2.ID})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cats))
	assert.Contains(t, cats, cat)
	assert.Contains(t, cats, cat2)

	oldCat2 := cat2
	cat2.Name = "New Name 2"
	cat2, err = p.UpdateCategory(context.TODO(), cat2.ID, cat2)
	assert.NoError(t, err)
	assert.Equal(t, "New Name 2", cat2.Name)

	// Try to update using old lastUpdateTime
	oldCat2, err = p.UpdateCategory(context.TODO(), oldCat2.ID, oldCat2)
	assert.Error(t, err)
	assert.Equal(t, model.ErrConcurrentUpdate, err.(model.Error).Code)

	// Try to update non-existent category
	cat3 := cat2
	cat3.ID = 123456
	cat3, err = p.UpdateCategory(context.TODO(), cat3.ID, cat3)
	assert.Error(t, err)
	assert.Equal(t, model.ErrNotFound, err.(model.Error).Code)

	cs, err := p.SelectCategories(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
}

func TestUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)

	then := time.Now()
	in, err := model.NewUserIn("Some One", "someone@example.com", false, "an-issuer", "https://issuer.example.com/someone"+then.String())
	assert.NoError(t, err)
	// Retrieve non-existent user
	user, err := p.RetrieveUser(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, in.UserBody, user.UserBody)
	assert.LessOrEqual(t, then.Unix(), user.InsertTime.Unix())
	assert.LessOrEqual(t, then.Unix(), user.LastUpdateTime.Unix())

	// Retrieve existing user
	user, err = p.RetrieveUser(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, in.UserBody, user.UserBody)

}

func TestCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)

	_, err := p.SelectOneCollection(context.TODO(), 1)
	assert.Error(t, err)
	assert.Equal(t, persist.ErrNoRows, err)

	ci, err := model.NewCategoryIn("Test Category")
	assert.NoError(t, err)

	cat, err := p.InsertCategory(context.TODO(), ci)
	assert.NoError(t, err)
	assert.Equal(t, ci.Name, cat.Name)

	// then := time.Now().Truncate(0)
	in := model.NewCollectionIn("Collection 1", []uint32{cat.ID})
	out, err := p.InsertCollection(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, in.Name, out.Name)
	// assert.Equal(t, in.Categories, out.Categories)
	// assert.LessOrEqual(t, then, out.InsertTime)
	// assert.LessOrEqual(t, then, out.LastUpdateTime)

	coll, err := p.SelectOneCollection(context.TODO(), out.ID)
	assert.NoError(t, err)
	assert.Equal(t, out.Name, coll.Name)
	assert.Equal(t, out.Categories, coll.Categories)
	assert.Equal(t, out.InsertTime, coll.InsertTime)
	assert.Equal(t, out.LastUpdateTime, coll.InsertTime)

	// Try to create a collection with a non-existent category ID
	in = model.NewCollectionIn("Bad Collection", []uint32{cat.ID, 123456, 234567})
	out, err = p.InsertCollection(context.TODO(), in)
	assert.Error(t, err)
	t.Logf("Error: %#v", err)
}

func TestSequences(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}

	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)

	v, err := p.GetSequenceValue()
	assert.NoError(t, err)
	assert.Less(t, uint32(0), v)
	vs, err := p.GetMultipleSequenceValues(50)
	assert.NoError(t, err)
	assert.Equal(t, 50, len(vs))
	for i := range vs {
		assert.Less(t, v, vs[i])
		if i > 0 {
			assert.Less(t, vs[i-1], vs[i])
		}
	}
}
