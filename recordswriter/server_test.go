package main_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/utils"

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
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)

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
	content := `"household","given","surname","birthdate","birthplace"
				"H1","fred","flintstone","19 Mar 1900","Autaugaville, AL"
				"H1","wilma","slaghoople","Abt 1900","AL"
				"H2","barney","rubble","",""`
	recordsKey := "2020-05-30/2020-05-30T00:00:00.000000000Z"
	fullRecordsKey := fmt.Sprintf("/%d/%s", 1, recordsKey)
	w, err := bucket.NewWriter(ctx, fullRecordsKey, nil)
	assert.NoError(t, err)
	_, err = fmt.Fprint(w, content)
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)

	// Add a test category and test collection and test post for referential integrity
	testCategory := createTestCategory(ctx, t, p)
	defer deleteTestCategory(ctx, t, p, testCategory)
	testCollection := createTestCollection(ctx, t, p, testCategory.ID)
	defer deleteTestCollection(ctx, t, p, testCollection)

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
	assert.Equal(t, model.RecordsStatusToLoad, testPost.RecordsStatus)

	var post *model.Post
	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for default (empty)
		post, errors = testAPI.GetPost(ctx, testPost.ID)
		assert.Nil(t, errors)
		if post.RecordsStatus == model.RecordsStatusDefault {
			break
		}
		log.Printf("Waiting for recordswriter %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.RecordsStatusDefault, post.RecordsStatus, "Expected records status to be empty, got %s", post.RecordsStatus)

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

func createTestCategory(ctx context.Context, t *testing.T, p model.CategoryPersister) *model.Category {
	in, err := model.NewCategoryIn("Test")
	assert.NoError(t, err)
	created, e := p.InsertCategory(ctx, in)
	assert.Nil(t, e)
	return created
}

func deleteTestCategory(ctx context.Context, t *testing.T, p model.CategoryPersister, category *model.Category) {
	e := p.DeleteCategory(ctx, category.ID)
	assert.Nil(t, e)
}

func createTestCollection(ctx context.Context, t *testing.T, p model.CollectionPersister, categoryID uint32) *model.Collection {
	in := model.NewCollectionIn("Test", []uint32{categoryID})
	in.Fields = []model.CollectionField{
		{
			Header: "household",
		},
		{
			Header: "given",
		},
		{
			Header: "surname",
		},
		{
			Header: "birthdate",
		},
		{
			Header: "birthplace",
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
	created, e := p.InsertCollection(ctx, in)
	assert.Nil(t, e)
	return created
}

func deleteTestCollection(ctx context.Context, t *testing.T, p model.CollectionPersister, collection *model.Collection) {
	e := p.DeleteCollection(ctx, collection.ID)
	assert.Nil(t, e)
}
