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
	p := persist.NewPostgresPersister(db)
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
	testCategory := createTestCategory(t, p)
	defer deleteTestCategory(t, p, testCategory)
	testCollection := createTestCollection(t, p, testCategory.ID)
	defer deleteTestCollection(t, p, testCollection)
	testPost := createTestPost(t, p, testCollection.ID)
	defer deleteTestPost(t, p, testPost)
	// create records
	testRecords := createTestRecords(t, p, testPost.ID)
	defer deleteTestRecords(t, p, testRecords)
	// force post to draft status
	testPost.RecordsStatus = model.PostDraft
	testPost, err = p.UpdatePost(ctx, testPost.ID, *testPost)
	assert.NoError(t, err, "Error updating test post")
	assert.Equal(t, model.PostDraft, testPost.RecordsStatus, "Unexpected post recordsStatus")

	// Publish post
	testPost.RecordsStatus = model.PostPublished
	testPost, err = testAPI.UpdatePost(ctx, testPost.ID, *testPost)
	assert.NoError(t, err, "Error setting post to published")

	var post *model.Post
	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for Ready
		post, err = testAPI.GetPost(ctx, testPost.ID)
		assert.NoError(t, err)
		if post.RecordsStatus == model.PostPublished {
			break
		}
		log.Printf("Waiting for publisher %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.PostPublished, post.RecordsStatus, "Expected post to be Published, got %s", post.RecordsStatus)

	// search records by id
	for _, testRecord := range testRecords {
		searchID := strconv.Itoa(int(testRecord.ID))
		res, err := testAPI.SearchByID(ctx, searchID)
		assert.NoError(t, err)
		assert.Equal(t, searchID, res.ID, "Record not found")
		assert.Equal(t, testCollection.ID, res.CollectionID, "Collection not found")
	}

	// Unpublish post
	testPost, err = testAPI.GetPost(ctx, testPost.ID)
	testPost.RecordsStatus = model.PostDraft
	testPost, err = testAPI.UpdatePost(ctx, testPost.ID, *testPost)
	assert.NoError(t, err, "Error setting post back to draft")

	post = nil
	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for Draft
		post, err = testAPI.GetPost(ctx, testPost.ID)
		assert.NoError(t, err)
		if post.RecordsStatus == model.PostDraft {
			break
		}
		log.Printf("Waiting for publisher %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.PostDraft, post.RecordsStatus, "Expected post to be Draft, got %s", post.RecordsStatus)

	// verify records no longer searchable
	for _, testRecord := range testRecords {
		searchID := strconv.Itoa(int(testRecord.ID))
		_, err := testAPI.SearchByID(ctx, searchID)
		assert.Error(t, err)
		assert.IsType(t, &model.Errors{}, err)
		assert.Len(t, err.(*model.Errors).Errs(), 1)
		assert.Equal(t, model.ErrNotFound, err.(*model.Errors).Errs()[0].Code, "err.(*model.Errors).Errs()[0]: %#v", err.(*model.Errors).Errs()[0])
	}

	// delete post
	err = testAPI.DeletePost(ctx, testPost.ID)
	assert.NoError(t, err)
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
			IxRole:  "principal",
			IxField: "given",
		},
		{
			Header:  "surname",
			IxRole:  "principal",
			IxField: "surname",
		},
	}
	created, e := p.InsertCollection(context.TODO(), in)
	assert.NoError(t, e)
	return created
}

func deleteTestCollection(t *testing.T, p model.CollectionPersister, collection *model.Collection) {
	e := p.DeleteCollection(context.TODO(), collection.ID)
	assert.NoError(t, e)
}

func createTestPost(t *testing.T, p model.PostPersister, collectionID uint32) *model.Post {
	in := model.NewPostIn("Test", collectionID, "")
	created, e := p.InsertPost(context.TODO(), in)
	assert.NoError(t, e)
	return created
}

func deleteTestPost(t *testing.T, p model.PostPersister, post *model.Post) {
	e := p.DeletePost(context.TODO(), post.ID)
	assert.NoError(t, e)
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

func createTestRecords(t *testing.T, p model.RecordPersister, postID uint32) []model.Record {
	var records []model.Record
	for _, data := range recordData {
		in := model.NewRecordIn(data, postID)
		record, e := p.InsertRecord(context.TODO(), in)
		assert.NoError(t, e)
		records = append(records, *record)
	}
	return records
}

func deleteTestRecords(t *testing.T, p model.RecordPersister, records []model.Record) {
	for _, record := range records {
		e := p.DeleteRecord(context.TODO(), record.ID)
		assert.NoError(t, e)
	}
}
