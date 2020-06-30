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

func TestSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	db, err := postgres.Open(context.TODO(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
			err,
			os.Getenv("DATABASE_URL"),
		)
	}
	p := persist.NewPostgresPersister(db)
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		SettingsPersister(p)

	// Read settings
	settings, errors := testApi.GetSettings(context.TODO())
	assert.Nil(t, errors)
	assert.True(t, settings.InsertTime.IsZero())

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
	assert.NotEqual(t, settings.InsertTime, settings.LastUpdateTime)
}
