package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
	"gocloud.dev/postgres"
)

func TestPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()

	// create test api
	db, err := postgres.Open(ctx, os.Getenv("DATABASE_URL"))
	assert.NoError(t, err)
	p := persist.NewPostgresPersister("", db)
	testAPI, err := api.NewAPI()
	assert.NoError(t, err)
	defer testAPI.Close()
	testAPI = testAPI.
		QueueConfig("publisher", "amqp://guest:guest@localhost:35672/").
		ElasticsearchConfig("http://localhost:19200", nil).
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
	// force post to draft status
	testPost.RecordsStatus = api.PostDraft
	*testPost, err = p.UpdatePost(ctx, testPost.ID, *testPost)
	assert.Nil(t, err, "Error updating test post")
	assert.Equal(t, api.PostDraft, testPost.RecordsStatus, "Unexpected post recordsStatus")
	// create records
	testRecords, err := createTestRecords(p, testPost.ID)
	assert.Nil(t, err, "Error creating test records")
	defer deleteTestRecords(p, testRecords)

	// Update Post
	testPost.RecordsStatus = api.PostPublished
	testPost, errors := testAPI.UpdatePost(ctx, testPost.ID, *testPost)
	assert.Nil(t, errors, "Error setting post to published")

	var post *model.Post
	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for Ready
		post, errors = testAPI.GetPost(ctx, testPost.ID)
		assert.Nil(t, errors)
		if post.RecordsStatus == api.PostPublished {
			break
		}
		log.Printf("Waiting for publisher %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, api.PostPublished, post.RecordsStatus, "Expected post to be Published, got %s", post.RecordsStatus)

	// search records by id
	for _, testRecord := range testRecords {
		searchID := strconv.Itoa(int(testRecord.ID))
		res, err := testAPI.SearchByID(ctx, searchID)
		assert.Nil(t, err)
		assert.Equal(t, searchID, res.ID, "Record not found")
		assert.Equal(t, testCollection.ID, res.CollectionID, "Collection not found")
	}

	// delete post
	errors = testAPI.DeletePost(ctx, testPost.ID)
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

func createTestCollection(p model.CollectionPersister, categoryID uint32) (*model.Collection, error) {
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

func createTestPost(p model.PostPersister, collectionID uint32) (*model.Post, error) {
	in := model.NewPostIn("Test", collectionID, "test")
	created, err := p.InsertPost(context.TODO(), in)
	if err != nil {
		return nil, err
	}
	return &created, err
}

func deleteTestPost(p model.PostPersister, post *model.Post) error {
	return p.DeletePost(context.TODO(), post.ID)
}

var recordData = []map[string]string{
	{
		"given":   "Fred",
		"surname": "Flintstone",
	},
	{
		"given":   "Wilma",
		"surname": "Slaghoople",
	},
}

func createTestRecords(p model.RecordPersister, postID uint32) ([]model.Record, error) {
	var records []model.Record
	for _, data := range recordData {
		in := model.NewRecordIn(data, postID)
		record, err := p.InsertRecord(context.TODO(), in)
		if err != nil {
			return records, err
		}
		records = append(records, record)
	}
	return records, nil
}

func deleteTestRecords(p model.RecordPersister, records []model.Record) error {
	var err error
	for _, record := range records {
		if e := p.DeleteRecord(context.TODO(), record.ID); e != nil {
			err = e
		}
	}
	return err
}
