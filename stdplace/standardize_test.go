package stdplace_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/stdplace"
	"github.com/stretchr/testify/assert"
	"gocloud.dev/postgres"
)

func TestPlaceStandardize(t *testing.T) {
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
		doStandardizePlaceTests(t, p)
	}

}

func doStandardizePlaceTests(t *testing.T, p model.PlacePersister) {
	tests := []struct {
		text             string
		defaultContainer string
		id               uint32
		fullName         string
		errCode          model.ErrorCode
	}{
		{
			text:     "Alabama",
			fullName: "Alabama, United States",
			id:       1501,
		},
		{
			text:     "AL",
			fullName: "Alabama, United States",
			id:       1501,
		},
		{
			text:     "Autauga, AL",
			fullName: "Autauga, Alabama, United States",
			id:       1502,
		},
		{
			text:     "Fake, AL",
			fullName: "Fake, Alabama, United States",
			id:       1510,
		},
		{
			text:     "Bonita, AL",
			fullName: "Bonita, Autauga, Alabama, United States",
			id:       1507,
		},
		{
			text:     "Bonita, Fake, AL",
			fullName: "Bonita, Autauga, Alabama, United States",
			id:       1507,
		},
		{
			text:     "Billingsley, Foo, AL",
			fullName: "Billingsley, Autauga, Alabama, United States",
			id:       1506,
		},
		{
			text:     "Fake, Autauga, AL",
			fullName: "Booth, Autauga, Alabama, United States",
			id:       1508,
		},
		{
			text:     "Autauga, AL",
			fullName: "Autauga, Alabama, United States",
			id:       1502,
		},
		{
			text:     "Foo, AL",
			fullName: "Foo, Alabama, United States",
			id:       0,
		},
		{
			text:     "Foo, Bar, Autauga, AL",
			fullName: "Foo, Bar, Autauga, Alabama, United States",
			id:       0,
		},
		{
			text:     "Bethel Grove",
			fullName: "Bethel Grove, Autauga, Alabama, United States",
			id:       1505,
		},
		{
			text:    "Fake",
			errCode: model.ErrNotFound,
		},
		{
			text:             "Fake",
			defaultContainer: "Alabama, United States",
			fullName:         "Fake, Alabama, United States",
			id:               1510,
		},
	}

	ctx := context.TODO()
	ps, err := stdplace.NewStandardizer(ctx, p)
	assert.NoError(t, err, "NewStandardizer")

	for _, test := range tests {
		place, err := ps.Standardize(ctx, test.text, test.defaultContainer)
		if test.errCode != "" {
			assert.True(t, test.errCode.Matches(err), test.text)
		} else {
			assert.NoError(t, err, test.text)
			assert.Equal(t, test.fullName, place.FullName, test.text)
			assert.Equal(t, test.id, place.ID, test.text)
		}
	}
}
