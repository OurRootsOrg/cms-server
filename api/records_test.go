package api_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/utils"

	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
)

func TestRecords(t *testing.T) {
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
		doRecordsTests(t, p, p, p, p)
	}
	// TODO uncomment when record_household has been implemented for dynamodb
	//dynamoDBTableName := os.Getenv("DYNAMODB_TEST_TABLE_NAME")
	//if dynamoDBTableName != "" {
	//	config := aws.Config{
	//		Region:      aws.String("us-east-1"),
	//		Endpoint:    aws.String("http://localhost:18000"),
	//		DisableSSL:  aws.Bool(true),
	//		Credentials: credentials.NewStaticCredentials("ACCESS_KEY", "SECRET", ""),
	//	}
	//	sess, err := session.NewSession(&config)
	//	assert.NoError(t, err)
	//	p, err := dynamo.NewPersister(sess, dynamoDBTableName)
	//	assert.NoError(t, err)
	//	doRecordsTests(t, p, p, p, p)
	//}
}
func doRecordsTests(t *testing.T,
	catP model.CategoryPersister,
	colP model.CollectionPersister,
	postP model.PostPersister,
	recordP model.RecordPersister,
) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		CategoryPersister(catP).
		CollectionPersister(colP).
		PostPersister(postP).
		RecordPersister(recordP)

	// Add a test category and test collection and test post for referential integrity
	testCategory := createTestCategory(ctx, t, catP)
	defer deleteTestCategory(ctx, t, catP, testCategory)
	testCollection := createTestCollection(ctx, t, colP, testCategory.ID)
	defer deleteTestCollection(ctx, t, colP, testCollection)
	testPost := createTestPost(ctx, t, postP, testCollection.ID)
	defer deleteTestPost(ctx, t, postP, testPost)

	empty, err := testApi.GetRecordsForPost(ctx, testPost.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(empty.Records), "Expected empty slice, got %#v", empty)

	// Add a Record
	in := model.RecordIn{
		RecordBody: model.RecordBody{
			Data: map[string]string{"foo": "bar"},
		},
		Post: testPost.ID,
	}
	created, err := testApi.AddRecord(ctx, in)
	assert.NoError(t, err)
	defer testApi.DeleteRecord(ctx, created.ID)
	assert.Equal(t, in.Data, created.Data, "Expected Name to match")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, in.Post, created.Post)

	// Add with bad post reference
	in.Post = in.Post + 88
	_, err = testApi.AddRecord(ctx, in)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrBadReference, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])

	// GET /records should now return the created Record
	ret, err := testApi.GetRecordsForPost(ctx, testPost.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ret.Records))
	assert.Equal(t, *created, ret.Records[0])

	// GET many records should now return the created Record
	records, err := testApi.GetRecordsByID(ctx, []uint32{created.ID}, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, *created, records[0])

	// GET /records/{id} should now return the created Record
	ret2, err := testApi.GetRecord(ctx, false, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created, &ret2.Record)

	// Bad request - no post
	in.Post = 0
	_, err = testApi.AddRecord(ctx, in)
	assert.IsType(t, &api.Error{}, err)
	if assert.Len(t, err.(*api.Error).Errs(), 1, "err.(*api.Error).Errs(): %#v", err.(*api.Error).Errs()) {
		assert.Equal(t, err.(*api.Error).Errs()[0].Code, model.ErrRequired)
	}

	// Record not found
	_, err = testApi.GetRecord(ctx, false, created.ID+99)
	assert.Error(t, err)
	assert.IsType(t, &api.Error{}, err)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])

	// Update
	ret2.Data = map[string]string{"foo": "baz"}
	updated, err := testApi.UpdateRecord(ctx, ret2.ID, ret2.Record)
	assert.NoError(t, err)
	assert.Equal(t, ret2.ID, updated.ID)
	assert.Equal(t, ret2.Post, updated.Post)
	assert.Equal(t, ret2.Data, updated.Data, "Expected Name to match")

	// Update non-existent
	_, err = testApi.UpdateRecord(ctx, updated.ID+99, *updated)
	assert.Error(t, err)
	assert.IsType(t, &api.Error{}, err)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])

	// Update with bad post
	updated.Post = updated.Post + 99
	_, err = testApi.UpdateRecord(ctx, updated.ID, *updated)
	assert.Error(t, err)
	assert.IsType(t, &api.Error{}, err)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrBadReference, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])

	// Update with bad LastUpdateTime
	updated.Post = ret2.Post
	updated.LastUpdateTime = time.Now().Add(-time.Minute)
	_, err = testApi.UpdateRecord(ctx, updated.ID, *updated)
	if assert.Error(t, err) {
		assert.IsType(t, &api.Error{}, err)
		assert.Len(t, err.(*api.Error).Errs(), 1)
		assert.Equal(t, model.ErrConcurrentUpdate, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])
	}

	// DELETE
	err = testApi.DeleteRecord(ctx, updated.ID)
	assert.NoError(t, err)

	// Record Households

	emptyHouseholds, err := testApi.GetRecordHouseholdsForPost(ctx, testPost.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(emptyHouseholds), "Expected empty slice, got %#v", emptyHouseholds)

	// Add a couple of household members
	in = model.RecordIn{
		RecordBody: model.RecordBody{
			Data: map[string]string{"HouseholdNumber": "H1"},
		},
		Post: testPost.ID,
	}
	mbr1, err := testApi.AddRecord(ctx, in)
	assert.NoError(t, err)
	defer testApi.DeleteRecord(ctx, mbr1.ID)
	mbr2, err := testApi.AddRecord(ctx, in)
	assert.NoError(t, err)
	defer testApi.DeleteRecord(ctx, mbr2.ID)

	// Add a Record Household
	inHousehold := model.RecordHouseholdIn{
		Post:      testPost.ID,
		Household: "H1",
		Records:   model.Uint32Slice{mbr1.ID, mbr2.ID},
	}
	createdHousehold, err := testApi.AddRecordHousehold(ctx, inHousehold)
	assert.NoError(t, err)
	defer testApi.DeleteRecordsForPost(ctx, testPost.ID)
	assert.Equal(t, inHousehold.Records, createdHousehold.Records, "Expected record ids to match")
	assert.Equal(t, inHousehold.Post, createdHousehold.Post)
	assert.Equal(t, inHousehold.Household, createdHousehold.Household)

	// GET record households should now return the created Record Households
	retHouseholds, err := testApi.GetRecordHouseholdsForPost(ctx, testPost.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(retHouseholds))
	assert.Equal(t, *createdHousehold, retHouseholds[0])

	// GET record household should now return the created Record
	retHh, err := testApi.GetRecordHousehold(ctx, createdHousehold.Post, createdHousehold.Household)
	assert.NoError(t, err)
	assert.Equal(t, createdHousehold, retHh)

	// GET record for a household member with include household set should return all household members
	detail, err := testApi.GetRecord(ctx, true, mbr1.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(detail.Household))
	assert.Equal(t, mbr1.ID, detail.Household[0].ID)
	assert.Equal(t, mbr2.ID, detail.Household[1].ID)

	// Bad request - no post
	inHousehold.Post = 0
	_, err = testApi.AddRecordHousehold(ctx, inHousehold)
	assert.IsType(t, &api.Error{}, err)
	if assert.Len(t, err.(*api.Error).Errs(), 1, "err.(*api.Error).Errs(): %#v", err.(*api.Error).Errs()) {
		assert.Equal(t, err.(*api.Error).Errs()[0].Code, model.ErrRequired)
	}

	// Record Household not found
	_, err = testApi.GetRecordHousehold(ctx, createdHousehold.Post, createdHousehold.Household+"a")
	assert.Error(t, err)
	assert.IsType(t, &api.Error{}, err)
	assert.Len(t, err.(*api.Error).Errs(), 1)
	assert.Equal(t, model.ErrNotFound, err.(*api.Error).Errs()[0].Code, "err.(*api.Error).Errs()[0]: %#v", err.(*api.Error).Errs()[0])

	// DELETE
	err = testApi.DeleteRecordHouseholdsForPost(ctx, testPost.ID)
	assert.NoError(t, err)
}

func createTestPost(ctx context.Context, t *testing.T, p model.PostPersister, collectionID uint32) *model.Post {
	in := model.NewPostIn("Test", collectionID, "")
	created, e := p.InsertPost(ctx, in)
	assert.NoError(t, e)
	return created
}

func deleteTestPost(ctx context.Context, t *testing.T, p model.PostPersister, post *model.Post) {
	e := p.DeletePost(ctx, post.ID)
	assert.NoError(t, e)
}
