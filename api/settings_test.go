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

func TestSettings(t *testing.T) {
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
		doSettingsTests(t, p)
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
		log.Print(p)
		doSettingsTests(t, p)
	}
}
func doSettingsTests(t *testing.T,
	p model.SettingsPersister,
) {
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		SettingsPersister(p)

	// Read settings
	settings, errors := testApi.GetSettings(context.TODO())
	assert.Nil(t, errors)
	// assert.True(t, settings.InsertTime.IsZero())

	// Update settings
	settings.PostMetadata = []model.SettingsPostMetadata{
		{
			Name: "One",
			Type: "string",
		},
	}
	settings, errors = testApi.UpdateSettings(context.TODO(), *settings)
	assert.Nil(t, errors)
	assert.False(t, settings.InsertTime.IsZero())

	// Read settings again
	settings, errors = testApi.GetSettings(context.TODO())
	// t.Logf("settings: %#v", settings)
	assert.Nil(t, errors)
	assert.False(t, settings.InsertTime.IsZero())

	// Update settings again
	settings.PostMetadata = append(settings.PostMetadata, model.SettingsPostMetadata{
		Name: "Two",
		Type: "number",
	})
	settings, errors = testApi.UpdateSettings(context.TODO(), *settings)
	assert.Nil(t, errors)
	assert.Equal(t, 2, len(settings.PostMetadata))
	// assert.NotEqual(t, settings.InsertTime, settings.LastUpdateTime)
}
