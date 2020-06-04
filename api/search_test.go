package api_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/model"
	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()

	// create test api
	db, err := postgres.Open(context.TODO(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
			err,
			os.Getenv("DATABASE_URL"),
		)
	}
	p := persist.NewPostgresPersister("", db)
	assert.NoError(t, err)
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		CategoryPersister(p).
		CollectionPersister(p).
		PostPersister(p).
		RecordPersister(p).
		ElasticsearchConfig("http://localhost:19200")

	// Add a test category and test collection and test post and test records
	testCategory, err := createTestCategory(p)
	assert.Nil(t, err, "Error creating test category")
	defer deleteTestCategory(p, testCategory)
	testCollection, err := createTestCollection(p, testCategory.ID)
	assert.Nil(t, err, "Error creating test collection")
	defer deleteTestCollection(p, testCollection)
	testPost, err := createTestPost(p, testCollection.ID)
	assert.Nil(t, err, "Error creating test post")
	defer deleteTestPost(p, testPost)
	records, err := createTestRecords(p, testPost.ID)
	assert.Nil(t, err, "Error creating test records")
	defer deleteTestRecords(p, records)

	// index post
	err = testApi.IndexPost(ctx, testPost)
	assert.Nil(t, err, "Error indexing post")
	time.Sleep(5 * time.Second)

	// search by id
	res, errs := testApi.SearchByID(ctx, records[0].ID)
	assert.Nil(t, errs, "Error searching by id")
	assert.True(t, res["found"].(bool), "Record not found")

	// search
	res, errs = testApi.Search(ctx, api.SearchRequest{Given: "Fred"})
	assert.Nil(t, errs, "Error searching by id")
	assert.GreaterOrEqual(t,
		res["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64),
		float64(1),
		"Indexed record not found in search")
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

func createTestRecords(p model.RecordPersister, postID string) ([]model.Record, error) {
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
