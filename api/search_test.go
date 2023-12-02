package api_test

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/utils"

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
		doSearchTests(t, p, p, p, p, p, p)
	}
	// TODO implement
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
	//	doSearchTests(t, p, p, p, p, p)
	//}
}

func doSearchTests(t *testing.T,
	catP model.CategoryPersister,
	colP model.CollectionPersister,
	postP model.PostPersister,
	recordP model.RecordPersister,
	nameP model.NamePersister,
	societyP model.SocietyPersister,
) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	ctx = utils.AddSearchUserIDToContext(ctx, 1)
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		CategoryPersister(catP).
		CollectionPersister(colP).
		PostPersister(postP).
		RecordPersister(recordP).
		NamePersister(nameP).
		SocietyPersister(societyP).
		ElasticsearchConfig("http://localhost:19200", nil)

	// Add a test category and test collection and test post and test records
	testCategory := createTestCategory(ctx, t, catP)
	defer deleteTestCategory(ctx, t, catP, testCategory)
	testCollection := createTestCollection(ctx, t, colP, testCategory.ID)
	defer deleteTestCollection(ctx, t, colP, testCollection)
	testPost := createTestPost(ctx, t, postP, testCollection.ID)
	defer deleteTestPost(ctx, t, postP, testPost)
	records := createTestRecords(ctx, t, recordP, testPost.ID)
	defer deleteTestRecords(ctx, t, recordP, records)
	createTestHousehold(ctx, t, recordP, testPost.ID, records)
	defer deleteTestHousehold(ctx, t, recordP, testPost.ID)

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
	hit, errs := testApi.SearchByID(ctx, searchID, nil)
	assert.Nil(t, errs, "Error searching by id")
	assert.Equal(t, searchID, hit.ID)
	assert.Equal(t, model.PrincipalRole, hit.Person.Role)
	assert.Equal(t, "Fred Flintstone", hit.Person.Name)
	assert.Equal(t, testCollection.ID, hit.CollectionID)
	assert.Equal(t, testCollection.Name, hit.CollectionName)
	assert.Equal(t, 2, len(hit.Record))
	assert.Equal(t, "Given", hit.Record[0].Label)
	assert.Equal(t, "Fred", hit.Record[0].Value)
	assert.Equal(t, 3, len(hit.Household))
	assert.Equal(t, "Given", hit.Household[0][0].Label)
	assert.Equal(t, "Fred", hit.Household[0][0].Value)
	assert.Equal(t, "Wilma", hit.Household[1][0].Value)
	assert.Equal(t, "Pebbles", hit.Household[2][0].Value)

	// search
	res, errs := testApi.Search(ctx, &api.SearchRequest{
		SocietyID:             1,
		Given:                 "Fred",
		CollectionPlace1:      "United States",
		CollectionPlace2Facet: true,
		Size:                  10,
	})
	assert.Nil(t, errs, "Error searching")
	assert.GreaterOrEqual(t, res.Total, 1)
	assert.GreaterOrEqual(t, len(res.Hits), 1)
	assert.Equal(t, "Fred Flintstone", res.Hits[0].Person.Name)
	assert.Equal(t, testCollection.ID, res.Hits[0].CollectionID)
	assert.Equal(t, testCollection.Name, res.Hits[0].CollectionName)
	assert.Nil(t, res.Hits[0].Record)
	assert.Equal(t, 1, len(res.Facets))
	assert.Equal(t, 1, len(res.Facets["collectionPlace2"].Buckets))
	assert.Equal(t, "Iowa", res.Facets["collectionPlace2"].Buckets[0].Label)
	assert.Equal(t, 1, res.Facets["collectionPlace2"].Buckets[0].Count)

	// search with relative
	res, errs = testApi.Search(ctx, &api.SearchRequest{
		SocietyID:            1,
		Given:                "Wilma",
		GivenFuzziness:       api.FuzzyNameExact,
		SpouseGiven:          "Fred",
		SpouseGivenFuzziness: api.FuzzyNameExact,
		Size:                 10,
	})
	assert.Nil(t, errs, "Error searching")
	assert.GreaterOrEqual(t, len(res.Hits), 1)
	assert.Equal(t, "Wilma Flintstone", res.Hits[0].Person.Name)

	res, errs = testApi.Search(ctx, &api.SearchRequest{
		SocietyID:            1,
		Given:                "Pebbles",
		GivenFuzziness:       api.FuzzyNameExact,
		FatherGiven:          "Fred",
		FatherGivenFuzziness: api.FuzzyNameExact,
		MotherGiven:          "Wilma",
		MotherGivenFuzziness: api.FuzzyNameExact,
		Size:                 10,
	})
	assert.Nil(t, errs, "Error searching")
	assert.GreaterOrEqual(t, len(res.Hits), 1)
	assert.Equal(t, "Pebbles Flintstone", res.Hits[0].Person.Name)

	res, errs = testApi.Search(ctx, &api.SearchRequest{
		SocietyID:            1,
		Given:                "Pebbles",
		GivenFuzziness:       api.FuzzyNameExact,
		FatherGiven:          "Imposter",
		FatherGivenFuzziness: api.FuzzyNameExact,
		Size:                 10,
	})
	assert.Nil(t, errs, "Error searching")
	assert.Equal(t, len(res.Hits), 0)
}

var recordData = []map[string]string{
	{
		"Given":           "Fred",
		"Surname":         "Flintstone",
		"HouseholdNumber": "H1",
		"RelToHead":       "HEAD",
		"Gender":          "Male",
	},
	{
		"Given":           "Wilma",
		"Surname":         "Flintstone",
		"HouseholdNumber": "H1",
		"RelToHead":       "Wife",
		"Gender":          "Female",
	},
	{
		"Given":           "Pebbles",
		"Surname":         "Flintstone",
		"HouseholdNumber": "H1",
		"RelToHead":       "Child",
		"Gender":          "Female",
	},
}

func createTestRecords(ctx context.Context, t *testing.T, p model.RecordPersister, postID uint32) []model.Record {
	var records []model.Record
	for _, data := range recordData {
		in := model.NewRecordIn(data, postID)
		record, e := p.InsertRecord(ctx, in)
		assert.Nil(t, e)
		records = append(records, *record)
	}
	return records
}

func deleteTestRecords(ctx context.Context, t *testing.T, p model.RecordPersister, records []model.Record) {
	for _, record := range records {
		e := p.DeleteRecord(ctx, record.ID)
		assert.Nil(t, e)
	}
}

func createTestHousehold(ctx context.Context, t *testing.T, p model.RecordPersister, postID uint32, records []model.Record) {
	var recordIDs []uint32
	for _, record := range records {
		recordIDs = append(recordIDs, record.ID)
	}
	inHousehold := model.RecordHouseholdIn{
		Post:      postID,
		Household: "H1",
		Records:   recordIDs,
	}
	_, e := p.InsertRecordHousehold(ctx, inHousehold)
	assert.Nil(t, e)
}

func deleteTestHousehold(ctx context.Context, t *testing.T, p model.RecordPersister, postID uint32) {
	e := p.DeleteRecordHouseholdsForPost(ctx, postID)
	assert.Nil(t, e)
}
