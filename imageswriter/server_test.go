package main

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"gocloud.dev/blob"
	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/stretchr/testify/assert"
)

func TestImagesWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()

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

	// write a zip file to a bucket
	bucket, err := testAPI.OpenBucket(ctx)
	assert.NoError(t, err)
	defer bucket.Close()

	zipBytes, err := ioutil.ReadFile("testdata/test.zip")
	assert.NoError(t, err)
	t.Logf("len(zipBytes): %d\n", len(zipBytes))
	ra := bytes.NewReader(zipBytes)
	assert.NoError(t, err)
	zr, err := zip.NewReader(ra, int64(len(zipBytes)))
	assert.NoError(t, err)
	zipNames := make(map[string]bool)
	for _, f := range zr.File {
		t.Logf("file name: %s\n", f.Name)
		zipNames[f.Name] = true
	}

	// write an object
	imagesKey := "/2020-08-10/2020-08-10T00:00:00.000000000Z"
	w, err := bucket.NewWriter(ctx, imagesKey+".zip", nil)
	assert.NoError(t, err)
	_, err = io.Copy(w, bytes.NewBuffer(zipBytes))
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)

	// Add a test category and test collection and test post for referential integrity
	testCategory := createTestCategory(t, p)
	defer deleteTestCategory(t, p, testCategory)
	testCollection := createTestCollection(t, p, testCategory.ID)
	defer deleteTestCollection(t, p, testCollection)

	// Add a Post
	in := model.PostIn{
		PostBody: model.PostBody{
			Name:      "Test Post",
			ImagesKey: imagesKey,
		},
		Collection: testCollection.ID,
	}
	log.Printf("[DEBUG] Adding post %#v", in)
	testPost, errors := testAPI.AddPost(ctx, in)
	assert.Nil(t, errors)

	var post *model.Post
	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for Draft
		post, errors = testAPI.GetPost(ctx, testPost.ID)
		assert.Nil(t, errors)
		if post.ImagesStatus == model.PostDraft {
			break
		}
		log.Printf("Waiting for imageswriter %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.PostDraft, post.ImagesStatus, "Expected post to be Draft, got %s", post.ImagesStatus)

	prefix := imagesKey + "/"
	// read images for post
	li := bucket.List(&blob.ListOptions{
		Prefix: prefix,
	})
	for {
		obj, err := li.Next(ctx)
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(obj.Key, prefix))
		suffix := strings.TrimPrefix(obj.Key, prefix)
		assert.True(t, zipNames[suffix])
	}

	// delete post
	errors = testAPI.DeletePost(ctx, testPost.ID)
	assert.Nil(t, errors)

	// images should be removed
	// images, errors = testAPI.GetImagesForPost(ctx, testPost.ID)
	// assert.Nil(t, errors)
	// assert.Equal(t, 0, len(images.Images), "Expected empty slice, got %#v", images)
}

func createTestCategory(t *testing.T, p model.CategoryPersister) *model.Category {
	in, err := model.NewCategoryIn("Test")
	assert.NoError(t, err)
	created, e := p.InsertCategory(context.TODO(), in)
	assert.Nil(t, e)
	return created
}

func deleteTestCategory(t *testing.T, p model.CategoryPersister, category *model.Category) {
	e := p.DeleteCategory(context.TODO(), category.ID)
	assert.Nil(t, e)
}

func createTestCollection(t *testing.T, p model.CollectionPersister, categoryID uint32) *model.Collection {
	in := model.NewCollectionIn("Test", []uint32{categoryID})
	created, e := p.InsertCollection(context.TODO(), in)
	assert.Nil(t, e)
	return created
}

func deleteTestCollection(t *testing.T, p model.CollectionPersister, collection *model.Collection) {
	e := p.DeleteCollection(context.TODO(), collection.ID)
	assert.Nil(t, e)
}
