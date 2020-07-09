package api_test

import (
	"context"
	"log"
	"os"
	"strconv"
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
	p := persist.NewPostgresPersister(db)
	assert.NoError(t, err)
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		CategoryPersister(p).
		CollectionPersister(p).
		PostPersister(p).
		RecordPersister(p).
		ElasticsearchConfig("http://localhost:19200", nil)

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
	time.Sleep(1 * time.Second)
	defer func() {
		for _, record := range records {
			_ = testApi.SearchDeleteByID(ctx, strconv.Itoa(int(record.ID)))
		}
	}()

	// search by id
	searchID := strconv.Itoa(int(records[0].ID))
	hit, errs := testApi.SearchByID(ctx, searchID)
	assert.Nil(t, errs, "Error searching by id")
	assert.Equal(t, searchID, hit.ID)
	assert.Equal(t, "principal", hit.Person.Role)
	assert.Equal(t, "Fred Flintstone", hit.Person.Name)
	assert.Equal(t, testCollection.ID, hit.CollectionID)
	assert.Equal(t, testCollection.Name, hit.CollectionName)
	assert.Equal(t, 2, len(hit.Record))
	assert.Equal(t, "Given", hit.Record[0].Label)
	assert.Equal(t, "Fred", hit.Record[0].Value)

	// search
	res, errs := testApi.Search(ctx, &api.SearchRequest{Given: "Fred"})
	assert.Nil(t, errs, "Error searching by id")
	assert.GreaterOrEqual(t, res.Total, 1)
	assert.GreaterOrEqual(t, len(res.Hits), 1)
	assert.Equal(t, "Fred Flintstone", res.Hits[0].Person.Name)
	assert.Equal(t, testCollection.ID, res.Hits[0].CollectionID)
	assert.Equal(t, testCollection.Name, res.Hits[0].CollectionName)
	assert.Nil(t, res.Hits[0].Record)
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
		records = append(records, *record)
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
