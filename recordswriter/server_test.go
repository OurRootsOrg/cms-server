package main_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/stretchr/testify/assert"
)

func TestRecordsWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()

	// create test api
	db, err := postgres.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
			err,
			os.Getenv("DATABASE_URL"),
		)
	}
	p := persist.NewPostgresPersister(db)
	testAPI, err := api.NewAPI()
	assert.NoError(t, err)
	defer testAPI.Close()
	testAPI = testAPI.
		BlobStoreConfig("us-east-1", "127.0.0.1:19000", "minioaccess", "miniosecret", "testbucket", true).
		QueueConfig("recordswriter", "amqp://guest:guest@localhost:35672/").
		QueueConfig("imageswriter", "amqp://guest:guest@localhost:35672/").
		CategoryPersister(p).
		CollectionPersister(p).
		PostPersister(p).
		RecordPersister(p)

	// write an object to a bucket
	bucket, err := testAPI.OpenBucket(ctx, false)
	assert.NoError(t, err)
	defer bucket.Close()

	// write an object
	content := `[{"household":"H1","given":"fred","surname":"flintstone","birthdate":"19 Mar 1900","birthplace":"Autaugaville, AL"},
				 {"household":"H1","given":"wilma","surname":"slaghoople","birthdate":"Abt 1900","birthplace":"AL"},
				 {"household":"H2","given":"barney","surname":"rubble"}]`
	recordsKey := "/2020-05-30/2020-05-30T00:00:00.000000000Z"
	w, err := bucket.NewWriter(ctx, recordsKey, nil)
	assert.NoError(t, err)
	_, err = fmt.Fprint(w, content)
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)

	// Add a test category and test collection and test post for referential integrity
	testCategory := createTestCategory(t, p)
	defer deleteTestCategory(t, p, testCategory)
	testCollection := createTestCollection(t, p, testCategory.ID)
	defer deleteTestCollection(t, p, testCollection)

	// Add a Post
	in := model.PostIn{
		PostBody: model.PostBody{
			Name:       "Test Post",
			RecordsKey: recordsKey,
		},
		Collection: testCollection.ID,
	}
	testPost, errors := testAPI.AddPost(ctx, in)
	assert.Nil(t, errors)

	var post *model.Post
	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for Draft
		post, errors = testAPI.GetPost(ctx, testPost.ID)
		assert.Nil(t, errors)
		if post.RecordsStatus == model.PostDraft {
			break
		}
		log.Printf("Waiting for recordswriter %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.PostDraft, post.RecordsStatus, "Expected post to be Draft, got %s", post.RecordsStatus)

	// read records for post
	records, errors := testAPI.GetRecordsForPost(ctx, testPost.ID)
	assert.Nil(t, errors)
	assert.Equal(t, 3, len(records.Records), "Expected three records, got %#v", records)

	// check standardization results
	fredIx := -1
	wilmaIx := -1
	for ix, record := range records.Records {
		if record.Data["given"] == "fred" {
			fredIx = ix
		}
		if record.Data["given"] == "wilma" {
			wilmaIx = ix
		}
	}
	assert.GreaterOrEqual(t, fredIx, 0)
	assert.GreaterOrEqual(t, wilmaIx, 0)
	assert.Equal(t, "19000319", records.Records[fredIx].Data["birthdate_std"])
	assert.Equal(t, "19000000,18990101-19011231", records.Records[wilmaIx].Data["birthdate_std"])
	assert.Equal(t, "Autaugaville, Autauga, Alabama, United States", records.Records[fredIx].Data["birthplace_std"])
	assert.Equal(t, "Alabama, United States", records.Records[wilmaIx].Data["birthplace_std"])

	// read households for post
	recordHouseholds, errors := testAPI.GetRecordHouseholdsForPost(ctx, testPost.ID)
	assert.Nil(t, errors)
	assert.Equal(t, 2, len(recordHouseholds), "Expected two households, got %#v", recordHouseholds)
	flintstoneIx := -1
	rubbleIx := -1
	for ix, recordHousehold := range recordHouseholds {
		if recordHousehold.Household == "H1" {
			flintstoneIx = ix
		}
		if recordHousehold.Household == "H2" {
			rubbleIx = ix
		}
	}
	assert.Equal(t, 2, len(recordHouseholds[flintstoneIx].Records))
	assert.Equal(t, 1, len(recordHouseholds[rubbleIx].Records))

	// delete post
	errors = testAPI.DeletePost(ctx, testPost.ID)
	assert.Nil(t, errors)

	// records should be removed
	records, errors = testAPI.GetRecordsForPost(ctx, testPost.ID)
	assert.Nil(t, errors)
	assert.Equal(t, 0, len(records.Records), "Expected empty slice, got %#v", records)
}

func createTestCategory(t *testing.T, p model.CategoryPersister) *model.Category {
	in, err := model.NewCategoryIn("Test")
	assert.NoError(t, err)
	created, e := p.InsertCategory(context.TODO(), in)
	assert.Nil(t, e)
	return created
}

func deleteTestCategory(t *testing.T, p model.CategoryPersister, category *model.Category) {
	e := p.DeleteCategory(context.TODO(), category.ID)
	assert.Nil(t, e)
}

func createTestCollection(t *testing.T, p model.CollectionPersister, categoryID uint32) *model.Collection {
	in := model.NewCollectionIn("Test", []uint32{categoryID})
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
		{
			Header:  "birthdate",
			IxRole:  "principal",
			IxField: "birthDate",
		},
		{
			Header:  "birthplace",
			IxRole:  "principal",
			IxField: "birthPlace",
		},
	}
	in.HouseholdNumberHeader = "household"
	created, e := p.InsertCollection(context.TODO(), in)
	assert.Nil(t, e)
	return created
}

func deleteTestCollection(t *testing.T, p model.CollectionPersister, collection *model.Collection) {
	e := p.DeleteCollection(context.TODO(), collection.ID)
	assert.Nil(t, e)
}
