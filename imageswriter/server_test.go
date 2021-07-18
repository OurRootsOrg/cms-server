package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/utils"

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
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
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
	bucket, err := testAPI.OpenBucket(ctx, false)
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
	imagesKey := "2020-08-10/2020-08-10T00:00:00.000000000Z"
	fullImagesKey := fmt.Sprintf("/%d/%s", 1, imagesKey)
	w, err := bucket.NewWriter(ctx, fullImagesKey, nil)
	assert.NoError(t, err)
	_, err = io.Copy(w, bytes.NewBuffer(zipBytes))
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)
	log.Printf("[DEBUG] Wrote object %s", imagesKey)

	// write a second object
	imagesKey2 := "2020-08-11/2020-08-11T00:00:00.000000000Z"
	fullImagesKey2 := fmt.Sprintf("/%d/%s", 1, imagesKey2)
	w, err = bucket.NewWriter(ctx, fullImagesKey2, nil)
	assert.NoError(t, err)
	_, err = io.Copy(w, bytes.NewBuffer(zipBytes))
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)
	log.Printf("[DEBUG] Wrote object %s", imagesKey2)

	// Add a test category and test collection and test post for referential integrity
	testCategory := createTestCategory(ctx, t, p)
	defer deleteTestCategory(ctx, t, p, testCategory)
	testCollection := createTestCollection(ctx, t, p, testCategory.ID)
	defer deleteTestCollection(ctx, t, p, testCollection)

	// Add a Post
	in := model.PostIn{
		PostBody: model.PostBody{
			Name:       "Test Post",
			ImagesKeys: model.StringSet{imagesKey},
		},
		Collection: testCollection.ID,
	}
	log.Printf("[DEBUG] Adding post %#v", in)
	testPost, errors := testAPI.AddPost(ctx, in)
	assert.Nil(t, errors)
	assert.Equal(t, model.ImagesStatusToLoad, testPost.ImagesStatus)

	var post *model.Post
	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for Default (empty)
		post, errors = testAPI.GetPost(ctx, testPost.ID)
		assert.Nil(t, errors)
		if post.ImagesStatus == model.ImagesStatusDefault {
			break
		}
		log.Printf("Waiting for imageswriter %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.ImagesStatusDefault, post.ImagesStatus, "Expected images status to be empty, got %s", post.ImagesStatus)

	prefix := fmt.Sprintf(api.ImagesPrefix, post.ID)
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
		suffix = strings.TrimSuffix(suffix, model.ImageDimensionsSuffix)
		suffix = strings.TrimSuffix(suffix, model.ImageThumbnailSuffix)
		assert.True(t, zipNames[suffix], suffix)
	}

	in = model.PostIn{
		PostBody: model.PostBody{
			Name:       "Test Post 2",
			ImagesKeys: model.StringSet{imagesKey, imagesKey2},
		},
		Collection: testCollection.ID,
	}
	log.Printf("[DEBUG] Adding post %#v", in)
	testPost2, errors := testAPI.AddPost(ctx, in)
	assert.Nil(t, errors)
	assert.Equal(t, model.ImagesStatusToLoad, testPost2.ImagesStatus)

	// wait up to 10 seconds
	for i := 0; i < 10; i++ {
		// read post and look for Default (empty)
		post, errors = testAPI.GetPost(ctx, testPost2.ID)
		assert.Nil(t, errors)
		if post.ImagesStatus == model.ImagesStatusDefault {
			break
		}
		log.Printf("Waiting for imageswriter %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.ImagesStatusDefault, post.ImagesStatus, "Expected images status to be empty, got %s", post.ImagesStatus)

	// delete posts
	errors = testAPI.DeletePost(ctx, testPost.ID)
	assert.Nil(t, errors)
	errors = testAPI.DeletePost(ctx, testPost2.ID)
	assert.Nil(t, errors)
}

func TestPostImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)

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

	zipBytes, err := ioutil.ReadFile("testdata/test.zip")
	const zipImageWidth = 212
	assert.NoError(t, err)
	t.Logf("len(zipBytes): %d\n", len(zipBytes))
	ra := bytes.NewReader(zipBytes)
	assert.NoError(t, err)
	zr, err := zip.NewReader(ra, int64(len(zipBytes)))
	assert.NoError(t, err)
	zipNames := make(map[string]bool)
	for _, f := range zr.File {
		t.Logf("file name: %s mode.IsDir=%t \n", f.Name, f.Mode().IsDir())
		if !f.Mode().IsDir() {
			zipNames[f.Name] = true
		}
	}

	// write a zip file to a bucket
	// make the request
	contentRequest, errs := testAPI.PostContentRequest(ctx, api.ContentRequest{ContentType: "application/zip"})
	assert.Nil(t, errs)

	// post the content
	client := &http.Client{}
	req, err := http.NewRequest("PUT", contentRequest.SignedURL, bytes.NewReader(zipBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/zip")
	res, err := client.Do(req)
	assert.Equal(t, 200, res.StatusCode)

	// Add a test category and test collection and test post for referential integrity
	testCategory := createTestCategory(ctx, t, p)
	defer deleteTestCategory(ctx, t, p, testCategory)
	testCollection := createTestCollection(ctx, t, p, testCategory.ID)
	defer deleteTestCollection(ctx, t, p, testCollection)

	// Add a Post
	in := model.PostIn{
		PostBody: model.PostBody{
			Name:       "Test Post",
			ImagesKeys: model.StringSet{contentRequest.Key},
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
		if post.ImagesStatus == model.ImagesStatusDefault {
			break
		}
		log.Printf("Waiting for imageswriter %d\n", i)
		time.Sleep(1 * time.Second)
	}
	assert.Equal(t, model.ImagesStatusDefault, post.ImagesStatus, "Expected images status to be empty, got %s", post.ImagesStatus)

	// give some additional time for thumbnails to be generated
	time.Sleep(3 * time.Second)

	// read images for post
	for name := range zipNames {
		// read image
		imageMetadata, errors := testAPI.GetPostImage(ctx, testPost.ID, name, false, 60)
		assert.Nil(t, errors, name)
		assert.Equal(t, zipImageWidth, imageMetadata.Width)
		resp, err := http.Get(imageMetadata.URL)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		fileBytes, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)

		thumbMetadata, errors := testAPI.GetPostImage(ctx, testPost.ID, name, true, 60)
		assert.Nil(t, errors, name)
		assert.Equal(t, model.ImageThumbnailWidth, thumbMetadata.Width)
		resp, err = http.Get(thumbMetadata.URL)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		fileBytes, err = ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)

		if !strings.HasSuffix(name, "/") {
			// Don't check directories
			assert.Less(t, 0, len(fileBytes), "Length of %s should be greater than 0", name)
		}
	}

	// delete post
	errors = testAPI.DeletePost(ctx, testPost.ID)
	assert.Nil(t, errors)
}

func createTestCategory(ctx context.Context, t *testing.T, p model.CategoryPersister) *model.Category {
	in, err := model.NewCategoryIn("Test")
	assert.NoError(t, err)
	created, e := p.InsertCategory(ctx, in)
	assert.Nil(t, e)
	return created
}

func deleteTestCategory(ctx context.Context, t *testing.T, p model.CategoryPersister, category *model.Category) {
	e := p.DeleteCategory(ctx, category.ID)
	assert.Nil(t, e)
}

func createTestCollection(ctx context.Context, t *testing.T, p model.CollectionPersister, categoryID uint32) *model.Collection {
	in := model.NewCollectionIn("Test", []uint32{categoryID})
	created, e := p.InsertCollection(ctx, in)
	assert.Nil(t, e)
	return created
}

func deleteTestCollection(ctx context.Context, t *testing.T, p model.CollectionPersister, collection *model.Collection) {
	e := p.DeleteCollection(ctx, collection.ID)
	assert.Nil(t, e)
}
