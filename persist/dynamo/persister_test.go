package dynamo_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist/dynamo"
	"github.com/stretchr/testify/assert"
)

func TestPersister(t *testing.T) {
	config := aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("http://localhost:18000"),
		DisableSSL:  aws.Bool(true),
		Credentials: credentials.NewStaticCredentials("ACCESS_KEY", "SECRET", ""),
	}
	sess, err := session.NewSession(&config)
	assert.NoError(t, err)
	p, err := dynamo.NewPersister(sess, "cms_test")
	assert.NoError(t, err)
	_, err = p.SelectOneCategory(context.TODO(), 1)
	assert.Error(t, err)
	assert.IsType(t, model.Error{}, err)
	assert.Equal(t, model.ErrNotFound, err.(model.Error).Code)
	assert.Equal(t, "1", err.(model.Error).Params[0])

	ci := makeCategoryIn(t)
	// now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time

	cat, err := p.InsertCategory(context.TODO(), ci)
	assert.NoError(t, err)
	assert.Equal(t, ci.Name, cat.Name)

	c, err := p.SelectOneCategory(context.TODO(), cat.ID)
	assert.NoError(t, err)
	assert.Equal(t, cat, c)

	// Add another
	ci = makeCategoryIn(t)

	cat2, err := p.InsertCategory(context.TODO(), ci)
	assert.NoError(t, err)
	assert.Equal(t, ci.Name, cat2.Name)

	// Get both in one call
	cats, err := p.SelectCategoriesByID(context.TODO(), []uint32{cat.ID, cat2.ID})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cats))
	assert.Contains(t, cats, cat)
	assert.Contains(t, cats, cat2)

	oldCat2 := cat2
	cat2.Name = "New Name 2"
	cat2, err = p.UpdateCategory(context.TODO(), cat2.ID, cat2)
	assert.NoError(t, err)
	assert.Equal(t, "New Name 2", cat2.Name)

	// Try to update using old lastUpdateTime
	oldCat2, err = p.UpdateCategory(context.TODO(), oldCat2.ID, oldCat2)
	assert.Error(t, err)
	assert.Equal(t, model.ErrConcurrentUpdate, err.(model.Error).Code)

	// Try to update non-existent category
	cat3 := cat2
	cat3.ID = 123456
	cat3, err = p.UpdateCategory(context.TODO(), cat3.ID, cat3)
	assert.Error(t, err)
	assert.Equal(t, model.ErrNotFound, err.(model.Error).Code)

	cs, err := p.SelectCategories(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, cs, 2)

	for _, c := range cs {
		err = p.DeleteCategory(context.TODO(), c.ID)
		assert.NoError(t, err)
	}
}

func makeCategoryIn(t *testing.T) model.CategoryIn {
	in, err := model.NewCategoryIn("Test Category")
	assert.NoError(t, err)
	return in
}
func makeCategory(t *testing.T) model.Category {
	now := time.Now()
	in := model.Category{
		ID:             33,
		CategoryBody:   makeCategoryIn(t).CategoryBody,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return in
}
