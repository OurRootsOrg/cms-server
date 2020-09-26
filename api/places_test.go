package api_test

import (
	"context"
	"log"
	"os"
	"testing"

	"gocloud.dev/postgres"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/persist/dynamo"
	"github.com/stretchr/testify/assert"
)

func TestPlaces(t *testing.T) {
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
		doPlaceTests(t, p)
	}
	dynamoDBTableName := os.Getenv("DYNAMODB_TEST_TABLE_NAME")
	if dynamoDBTableName != "" {
		config := aws.Config{
			Region:      aws.String("us-east-1"),
			Endpoint:    aws.String("http://localhost:18000"),
			DisableSSL:  aws.Bool(true),
			Credentials: credentials.NewStaticCredentials("ACCESS_KEY", "SECRET", ""),
		}
		sess, err := session.NewSession(&config)
		assert.NoError(t, err)
		p, err := dynamo.NewPersister(sess, dynamoDBTableName)
		assert.NoError(t, err)
		doPlaceTests(t, p)
	}
}

func doPlaceTests(t *testing.T, p model.PlacePersister) {
	ap, err := api.NewAPI()
	assert.NoError(t, err)
	defer ap.Close()
	testApi := ap.PlacePersister(p)
	result, err := testApi.GetPlacesByPrefix(context.TODO(), "Aut, ala", 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result), "Expected two places, got %#v", result)
	assert.Equal(t, "Autauga", result[0].Name)
	assert.Equal(t, "Autaugaville", result[1].Name)
}
