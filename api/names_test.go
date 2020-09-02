package api_test

import (
	"context"
	"log"
	"os"
	"testing"

	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/stretchr/testify/assert"
)

func TestNames(t *testing.T) {
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
		doNameTests(t, p)
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
	//	doPlaceTests(t, p)
	//}

}
func doNameTests(t *testing.T, p model.NamePersister) {
	ap, err := api.NewAPI()
	assert.NoError(t, err)
	defer ap.Close()
	testApi := ap.NamePersister(p)

	result, err := testApi.GetNameVariants(context.TODO(), model.GivenType, "fred")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(result.Variants))

	result, err = testApi.GetNameVariants(context.TODO(), model.SurnameType, "bedrock")
	assert.NoError(t, err)
	assert.Equal(t, 4, len(result.Variants))
}
