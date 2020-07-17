package api_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/persist/dynamo"
	"github.com/ourrootsorg/go-oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gocloud.dev/postgres"
	"golang.org/x/oauth2"
)

type mockProvider struct {
	mock.Mock
}

func (m *mockProvider) UserInfo(ctx context.Context, tokenSource oauth2.TokenSource) (*oidc.UserInfo, error) {
	rets := m.Called(ctx, tokenSource)
	return rets[0].(*oidc.UserInfo), rets.Error(1)
}

func TestUsers(t *testing.T) {
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
		doUserTests(t, p)
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
		doUserTests(t, p)
	}
}
func doUserTests(t *testing.T,
	p model.UserPersister,
) {
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.UserPersister(p)

	// model.User(&model.User{ID:"/users/1", UserIn:model.UserIn{UserBody:model.UserBody{Name:"<Unknown>", Email:"somebody@example.com", EmailConfirmed:true, Issuer:"https://flybynight.com", Subject:"user1", Enabled:true}}, InsertTime:time.Time{wall:0x1b4b8898, ext:63725606053, loc:(*time.Location)(0xc0000b4780)},
	expectedUser := model.User{
		ID: 1,
		UserBody: model.UserBody{
			Name:           "<Unknown>",
			Email:          "somebody@example.com",
			EmailConfirmed: true,
			Issuer:         "https://flybynight.com",
			Subject:        "user1",
			Enabled:        true,
		},
	}
	ctx := context.TODO()
	provider := mockProvider{}
	token := oidc.IDToken{
		Issuer:  "https://flybynight.com",
		Subject: "user1",
	}
	ui := oidc.UserInfo{
		Email:         "somebody@example.com",
		EmailVerified: true,
	}
	rawToken := "Abc"
	// First time through, not in DB or cache
	provider.On("UserInfo", ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: rawToken,
		TokenType:   "bearer",
	})).Once().Return(&ui, nil)

	user, errors := testApi.RetrieveUser(ctx, &provider, &token, rawToken)
	assert.Nil(t, errors)
	// assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.EmailConfirmed, user.EmailConfirmed)
	assert.Equal(t, expectedUser.Issuer, user.Issuer)
	assert.Equal(t, expectedUser.Subject, user.Subject)
	assert.Equal(t, expectedUser.Enabled, user.Enabled)
	// provider.AssertExpectations(t)

	// Second time through, in DB and cache
	user, errors = testApi.RetrieveUser(ctx, &provider, &token, rawToken)
	assert.Nil(t, errors)
	// assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.EmailConfirmed, user.EmailConfirmed)
	assert.Equal(t, expectedUser.Issuer, user.Issuer)
	assert.Equal(t, expectedUser.Subject, user.Subject)
	assert.Equal(t, expectedUser.Enabled, user.Enabled)
	// provider.AssertExpectations(t)

	// New API, in DB and not cache
	testApi, err = api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.UserPersister(p)

	provider.On("UserInfo", ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: rawToken,
		TokenType:   "bearer",
	})).Once().Return(&ui, nil)

	user, errors = testApi.RetrieveUser(ctx, &provider, &token, rawToken)
	assert.Nil(t, errors)
	// assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.EmailConfirmed, user.EmailConfirmed)
	assert.Equal(t, expectedUser.Issuer, user.Issuer)
	assert.Equal(t, expectedUser.Subject, user.Subject)
	assert.Equal(t, expectedUser.Enabled, user.Enabled)
	// provider.AssertExpectations(t)

}
