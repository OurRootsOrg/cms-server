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
	assert.Nil(t, err)
	p, err := dynamo.NewPersister(sess, table)
	assert.Nil(t, err)
	return p, func(t *testing.T) {
		t.Log("teardown test case")

		posts, err := p.SelectPosts(context.TODO())
		assert.NoError(t, err)
		for _, post := range posts {
			err = p.DeleteRecordsForPost(context.TODO(), post.ID)
			assert.NoError(t, err)

			err = p.DeletePost(context.TODO(), post.ID)
			assert.NoError(t, err)
			// Make sure it's gone
			_, err := p.SelectOnePost(context.TODO(), post.ID)
			assert.Error(t, err)
		}
		records, err := p.SelectRecords(context.TODO())
		assert.NoError(t, err)
		assert.Len(t, records, 0)

		colls, err := p.SelectCollections(context.TODO())
		assert.NoError(t, err)

		for _, c := range colls {
			err = p.DeleteCollection(context.TODO(), c.ID)
			assert.NoError(t, err)
			// Make sure it's gone
			_, err := p.SelectOneCollection(context.TODO(), c.ID)
			assert.Error(t, err)
		}

		cats, err := p.SelectCategories(context.TODO())
		assert.NoError(t, err)

		for _, c := range cats {
			err = p.DeleteCategory(context.TODO(), c.ID)
			assert.NoError(t, err)
			// Make sure it's gone
			_, err := p.SelectOneCategory(context.TODO(), c.ID)
			assert.Error(t, err)
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
	assert.IsType(t, &model.Error{}, err)
	assert.Equal(t, model.ErrNotFound, err.(*model.Error).Code)
	assert.Equal(t, "1", err.(*model.Error).Params[0])

	ci, err := model.NewCategoryIn("Test Category")
	assert.Nil(t, err)

	cat, e := p.InsertCategory(context.TODO(), ci)
	assert.Nil(t, e)
	assert.Equal(t, ci.Name, cat.Name)

	c, e := p.SelectOneCategory(context.TODO(), cat.ID)
	assert.Nil(t, e)
	assert.Equal(t, cat, c)

	// Add another
	ci, err = model.NewCategoryIn("Test Category 2")
	assert.Nil(t, err)

	cat2, e := p.InsertCategory(context.TODO(), ci)
	assert.Nil(t, e)
	assert.Equal(t, ci.Name, cat2.Name)

	// Get both in one call
	cats, e := p.SelectCategoriesByID(context.TODO(), []uint32{cat.ID, cat2.ID})
	assert.Nil(t, e)
	assert.Equal(t, 2, len(cats))
	assert.Contains(t, cats, *cat)
	assert.Contains(t, cats, *cat2)

	oldCat2 := cat2
	cat2.Name = "New Name 2"
	cat2, e = p.UpdateCategory(context.TODO(), cat2.ID, *cat2)
	assert.Nil(t, e)
	assert.Equal(t, "New Name 2", cat2.Name)

	// Try to update using old lastUpdateTime
	oldCat2, err = p.UpdateCategory(context.TODO(), oldCat2.ID, *oldCat2)
	assert.Error(t, err)
	assert.IsType(t, &model.Error{}, err)
	assert.Equal(t, model.ErrConcurrentUpdate, err.(*model.Error).Code)

	// Try to update non-existent category
	cat3 := cat2
	cat3.ID = 123456
	cat3, err = p.UpdateCategory(context.TODO(), cat3.ID, *cat3)
	assert.Error(t, err)
	assert.IsType(t, &model.Error{}, err)
	assert.Equal(t, model.ErrNotFound, err.(*model.Error).Code)

	cs, e := p.SelectCategories(context.TODO())
	assert.Nil(t, e)
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
	user, e := p.RetrieveUser(context.TODO(), in)
	assert.Nil(t, e)
	assert.Equal(t, in.UserBody, user.UserBody)
	assert.LessOrEqual(t, then.Unix(), user.InsertTime.Unix())
	assert.LessOrEqual(t, then.Unix(), user.LastUpdateTime.Unix())

	// Retrieve existing user
	user, e = p.RetrieveUser(context.TODO(), in)
	assert.Nil(t, e)
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
	assert.IsType(t, &model.Error{}, err)
	assert.Equal(t, model.ErrNotFound, err.(*model.Error).Code)

	ci, err := model.NewCategoryIn("Test Category")
	assert.Nil(t, err)

	cat, err := p.InsertCategory(context.TODO(), ci)
	assert.NoError(t, err)
	assert.Equal(t, ci.Name, cat.Name)

	// then := time.Now().Truncate(0)
	in := model.NewCollectionIn("Collection 1", []uint32{cat.ID})
	out, e := p.InsertCollection(context.TODO(), in)
	assert.Nil(t, e)
	assert.Equal(t, in.Name, out.Name)
	// assert.Equal(t, in.Categories, out.Categories)
	// assert.LessOrEqual(t, then, out.InsertTime)
	// assert.LessOrEqual(t, then, out.LastUpdateTime)

	coll, e := p.SelectOneCollection(context.TODO(), out.ID)
	assert.Nil(t, e)
	assert.Equal(t, out.Name, coll.Name)
	assert.Equal(t, out.Categories, coll.Categories)
	assert.Equal(t, out.InsertTime, coll.InsertTime)
	assert.Equal(t, out.LastUpdateTime, coll.InsertTime)

	// Try to create a collection with a non-existent category ID
	in = model.NewCollectionIn("Bad Collection", []uint32{cat.ID, 123456, 234567})
	out, e = p.InsertCollection(context.TODO(), in)
	assert.NotNil(t, e)
	t.Logf("Error: %#v", e)
}

func TestPost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)

	_, err := p.SelectOnePost(context.TODO(), 1)
	assert.Error(t, err)
	assert.IsType(t, &model.Error{}, err)
	assert.Equal(t, model.ErrNotFound, err.(*model.Error).Code)

	cati, err := model.NewCategoryIn("Test Category")
	assert.Nil(t, err)

	cat, err := p.InsertCategory(context.TODO(), cati)
	assert.NoError(t, err)
	assert.Equal(t, cati.Name, cat.Name)

	ci := model.NewCollectionIn("Test Collection", []uint32{cat.ID})

	coll, err := p.InsertCollection(context.TODO(), ci)
	assert.NoError(t, err)
	assert.Equal(t, ci.Name, coll.Name)

	// then := time.Now().Truncate(0)
	in := model.NewPostIn("Post 1", coll.ID, "")
	out, err := p.InsertPost(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, in.Name, out.Name)
	assert.Equal(t, in.Collection, out.Collection)
	assert.Equal(t, in.RecordsKey, out.RecordsKey)

	post, err := p.SelectOnePost(context.TODO(), out.ID)
	assert.NoError(t, err)
	assert.Equal(t, out.Name, post.Name)
	assert.Equal(t, out.Collection, post.Collection)
	assert.Equal(t, out.InsertTime, post.InsertTime)
	assert.Equal(t, out.LastUpdateTime, post.InsertTime)

	// Try to create a post with a non-existent collection ID
	in = model.NewPostIn("Bad Post", 123456, "")
	out, err = p.InsertPost(context.TODO(), in)
	assert.Error(t, err)
	// t.Logf("Error: %#v", err)
}

func TestRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)

	_, err := p.SelectOneRecord(context.TODO(), 1)
	assert.Error(t, err)
	assert.IsType(t, &model.Error{}, err)
	assert.Equal(t, model.ErrNotFound, err.(*model.Error).Code)

	cati, err := model.NewCategoryIn("Test Category")
	assert.Nil(t, err)

	cat, err := p.InsertCategory(context.TODO(), cati)
	assert.NoError(t, err)
	assert.Equal(t, cati.Name, cat.Name)

	ci := model.NewCollectionIn("Test Collection", []uint32{cat.ID})

	coll, e := p.InsertCollection(context.TODO(), ci)
	assert.Nil(t, e)
	assert.Equal(t, ci.Name, coll.Name)

	// then := time.Now().Truncate(0)
	postIn := model.NewPostIn("Post 1", coll.ID, "")
	post, e := p.InsertPost(context.TODO(), postIn)
	assert.Nil(t, e)
	assert.Equal(t, postIn.Name, post.Name)

	// then := time.Now().Truncate(0)
	recordIn := model.NewRecordIn(map[string]string{"test": "data"}, post.ID)
	out, e := p.InsertRecord(context.TODO(), recordIn)
	assert.Nil(t, e)
	assert.Equal(t, recordIn.Data, out.Data)
	assert.Equal(t, recordIn.Post, out.Post)

	record, e := p.SelectOneRecord(context.TODO(), out.ID)
	assert.Nil(t, e)
	assert.Equal(t, out.Data, record.Data)
	assert.Equal(t, out.Post, record.Post)
	assert.Equal(t, out.InsertTime, record.InsertTime)
	assert.Equal(t, out.LastUpdateTime, record.InsertTime)

	// Try to create a record with a non-existent  ID
	recordIn = model.NewRecordIn(map[string]string{"test": "Bad Record"}, 123456)
	out, err = p.InsertRecord(context.TODO(), recordIn)
	assert.Error(t, err)
	t.Logf("Error: %#v", e)
}

func TestSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)
	s, err := p.SelectSettings(context.TODO())
	assert.Error(t, err)
	assert.Nil(t, s)
	var in model.Settings
	in.PostMetadata = append(in.PostMetadata, model.SettingsPostMetadata{
		Name:    "First",
		Type:    "string",
		Tooltip: "Tooltip 1",
	})
	s, err = p.UpsertSettings(context.TODO(), in)
	assert.NoError(t, err)
	in = *s
	s, err = p.SelectSettings(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, in.PostMetadata, s.PostMetadata)
}

func TestSequences(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	p, teardown := setupTestCase(t)
	defer teardown(t)

	v, e := p.GetSequenceValue()
	assert.Nil(t, e)
	assert.Less(t, uint32(0), v)
	vs, e := p.GetMultipleSequenceValues(50)
	assert.Nil(t, e)
	assert.Equal(t, 50, len(vs))
	for i := range vs {
		assert.Less(t, v, vs[i])
		if i > 0 {
			assert.Less(t, vs[i-1], vs[i])
		}
	}
}
