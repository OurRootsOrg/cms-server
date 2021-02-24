package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ourrootsorg/cms-server/persist"
	"github.com/ourrootsorg/cms-server/persist/dynamo"
	"gocloud.dev/postgres"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestGetAllPosts(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	// Empty result
	cr := api.PostResult{}
	am.Result = &cr
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/societies/1/posts", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	var empty api.PostResult
	err := json.NewDecoder(response.Body).Decode(&empty)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 0, len(empty.Posts), "Expected empty slice, got %#v", empty)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])

	// Non-empty result
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	ci, _ := makePostIn(t, 0)
	cr = api.PostResult{
		Posts: []model.Post{
			{
				ID:             1,
				PostIn:         ci,
				InsertTime:     now,
				LastUpdateTime: now,
			},
		},
	}
	am.Result = &cr
	am.Errors = nil
	request, _ = http.NewRequest("GET", "/societies/1/posts", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var ret api.PostResult
	err = json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, 1, len(ret.Posts))
	assert.Equal(t, cr.Posts[0], ret.Posts[0])

	// error result
	am.Result = (*api.PostResult)(nil)
	am.Errors = api.NewError(assert.AnError)
	request, _ = http.NewRequest("GET", "/societies/1/posts", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, 500, response.Code)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var errRet []model.Error
	err = json.NewDecoder(response.Body).Decode(&errRet)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotNil(t, errRet)
	assert.Equal(t, 1, len(errRet))
	assert.Equal(t, am.Errors.(*api.Error).Errs(), errRet)
}
func TestGetPost(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	ci, _ := makePostIn(t, 0)
	post := &model.Post{
		ID:     1,
		PostIn: ci,
	}
	am.Result = post
	am.Errors = nil
	var ret model.Post

	request, _ := http.NewRequest("GET", "/societies/1/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	err := json.NewDecoder(response.Body).Decode(&ret)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, *post, ret)

	post = nil
	am.Result = post
	am.Errors = api.NewError(model.NewError(model.ErrNotFound, "1"))

	request, _ = http.NewRequest("GET", "/societies/1/posts/1", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var errRet []model.Error
	err = json.NewDecoder(response.Body).Decode(&errRet)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotNil(t, errRet)
	assert.Equal(t, 1, len(errRet))
	assert.Equal(t, am.Errors.(*api.Error).Errs(), errRet)
}

func TestGetPostImage(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	postID := 1
	imagePath := "foo/bar/image.jpg"

	url := "https://s3.example.com/mybucket" + fmt.Sprintf(api.ImagesPrefix, postID) + imagePath
	am.Result = &api.ImageMetadata{URL: url}
	am.Errors = nil

	request, _ := http.NewRequest("GET", "/societies/1/posts/1/images/"+imagePath, nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusTemporaryRedirect, response.Code)
	assert.Equal(t, url, response.Header().Get("Location"))

	request, _ = http.NewRequest("GET", "/societies/1/posts/1/images/"+imagePath+"?height=100", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusTemporaryRedirect, response.Code)
	assert.Equal(t, url, response.Header().Get("Location"))

	request, _ = http.NewRequest("GET", "/societies/1/posts/1/images/"+imagePath+"?noredirect=true", nil)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var metadata api.ImageMetadata
	err := json.NewDecoder(response.Body).Decode(&metadata)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, url, metadata.URL)
}

func TestPostPost(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makePostIn(t, 0)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	am.Result = &model.Post{
		ID:             1,
		PostIn:         in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.Errors = nil

	request, _ := http.NewRequest("POST", "/societies/1/posts", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Post
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.Equal(t, in, created.PostIn)
	assert.Equal(t, now, created.InsertTime)
	assert.Equal(t, now, created.LastUpdateTime)
	assert.NotEmpty(t, created.ID)
}

func TestPutPost(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	in, buf := makePostIn(t, 0)
	now := time.Now().Truncate(0) // Truncate(0) truncates monotonic time
	post := model.Post{
		ID:             1,
		PostIn:         in,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	am.Result = &post
	am.Errors = nil

	request, _ := http.NewRequest("PUT", "/societies/1/posts/1", buf)
	request.Header.Add("Content-Type", contentType)

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, "Response: %s", string(response.Body.Bytes()))
	assert.Contains(t, response.Result().Header, "Content-Type", "Should have Content-Type header")
	assert.Equal(t,
		contentType,
		response.Result().Header["Content-Type"][0])
	var created model.Post
	err := json.NewDecoder(response.Body).Decode(&created)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.Equal(t, in.Name, created.Name, "Expected Name to match")
	assert.Equal(t, in, created.PostIn)
	assert.Equal(t, now, created.InsertTime)
	assert.Equal(t, now, created.LastUpdateTime)
	assert.NotEmpty(t, created.ID)
}

func TestPutPostInvalidStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}

	// it doesn't matter whether the database is postgres or dynamodb; we're just testing invalid status updates
	var collectionPersister model.CollectionPersister
	var postPersister model.PostPersister
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
		collectionPersister = p
		postPersister = p
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
		collectionPersister = p
		postPersister = p
	}

	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.
		CollectionPersister(collectionPersister).
		PostPersister(postPersister)
	app := NewApp().API(testApi)
	app.authDisabled = true
	r := app.NewRouter()
	ctx := context.Background()

	// create a collection for referential integrity
	_, buf := makeCollectionIn(t, 0)
	request, _ := http.NewRequest("POST", "/societies/1/collections", buf)
	request.Header.Add("Content-Type", contentType)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	var collection model.Collection
	err = json.NewDecoder(response.Body).Decode(&collection)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotEmpty(t, collection.ID)
	defer collectionPersister.DeleteCollection(ctx, collection.ID)

	// create a post that we can try to update
	_, buf = makePostIn(t, collection.ID)
	request, _ = http.NewRequest("POST", "/societies/1/posts", buf)
	request.Header.Add("Content-Type", contentType)
	response = httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, "Response: %s", string(response.Body.Bytes()))
	post := &model.Post{}
	err = json.NewDecoder(response.Body).Decode(post)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	assert.NotEmpty(t, post.ID)
	defer postPersister.DeletePost(ctx, post.ID)

	// try various invalid post status updates
	postStatusTests := map[model.PostStatus][]model.PostStatus{
		model.PostStatusDraft: {
			model.PostStatusPublishing,
			model.PostStatusPublished,
			model.PostStatusPublishComplete,
			model.PostStatusUnpublishing,
			model.PostStatusUnpublishComplete,
			model.PostStatusPublishError,
			model.PostStatusUnpublishError,
			model.PostStatusError,
		},
		model.PostStatusPublishing: {
			model.PostStatusDraft,
			model.PostStatusPublished,
			model.PostStatusPublishComplete,
			model.PostStatusUnpublishing,
			model.PostStatusUnpublishComplete,
			model.PostStatusPublishError,
			model.PostStatusUnpublishError,
			model.PostStatusError,
		},
		model.PostStatusPublished: {
			model.PostStatusDraft,
			model.PostStatusPublishing,
			model.PostStatusPublishComplete,
			model.PostStatusUnpublishing,
			model.PostStatusUnpublishComplete,
			model.PostStatusPublishError,
			model.PostStatusUnpublishError,
			model.PostStatusError,
		},
		model.PostStatusError: {
			model.PostStatusDraft,
			model.PostStatusPublished,
			model.PostStatusPublishing,
			model.PostStatusPublishComplete,
			model.PostStatusUnpublishing,
			model.PostStatusUnpublishComplete,
			model.PostStatusPublishError,
			model.PostStatusUnpublishError,
		},
	}
	for currStatus, invalidStatuses := range postStatusTests {
		post.PostStatus = currStatus
		post.RecordsStatus = model.RecordsStatusDefault
		post.ImagesStatus = model.ImagesStatusDefault
		post, err = postPersister.UpdatePost(ctx, post.ID, *post)
		assert.NoError(t, err, fmt.Sprintf("updating post status %s", currStatus))

		for _, invalidStatus := range invalidStatuses {
			post.PostStatus = invalidStatus
			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			err := enc.Encode(post)
			if err != nil {
				t.Errorf("Error encoding PostIn: %v", err)
			}
			request, _ := http.NewRequest("PUT", fmt.Sprintf("/societies/1/posts/%d", post.ID), buf)
			request.Header.Add("Content-Type", contentType)
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)
			assert.Equal(t, http.StatusBadRequest, response.Code,
				fmt.Sprintf("updating post status from %s to %s expected bad request; response is: %d",
					currStatus, invalidStatus, response.Code))
		}
	}

	// try various invalid records status updates
	recordsStatusTests := map[model.RecordsStatus][]model.RecordsStatus{
		model.RecordsStatusDefault: {
			model.RecordsStatusToLoad,
			model.RecordsStatusLoading,
			model.RecordsStatusLoadComplete,
			model.RecordsStatusLoadError,
			model.RecordsStatusError,
		},
		model.RecordsStatusLoading: {
			model.RecordsStatusDefault,
			model.RecordsStatusToLoad,
			model.RecordsStatusLoadComplete,
			model.RecordsStatusLoadError,
			model.RecordsStatusError,
		},
		model.RecordsStatusError: {
			model.RecordsStatusDefault,
			model.RecordsStatusToLoad,
			model.RecordsStatusLoading,
			model.RecordsStatusLoadComplete,
			model.RecordsStatusLoadError,
		},
	}
	for currStatus, invalidStatuses := range recordsStatusTests {
		post.PostStatus = model.PostStatusDraft
		post.RecordsStatus = currStatus
		post.ImagesStatus = model.ImagesStatusDefault
		post, err = postPersister.UpdatePost(ctx, post.ID, *post)
		assert.NoError(t, err, fmt.Sprintf("updating records status %s", currStatus))

		for _, invalidStatus := range invalidStatuses {
			post.RecordsStatus = invalidStatus
			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			err := enc.Encode(post)
			if err != nil {
				t.Errorf("Error encoding PostIn: %v", err)
			}
			request, _ := http.NewRequest("PUT", fmt.Sprintf("/societies/1/posts/%d", post.ID), buf)
			request.Header.Add("Content-Type", contentType)
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)
			assert.Equal(t, http.StatusBadRequest, response.Code,
				fmt.Sprintf("updating records status from %s to %s expected bad request; response is: %d",
					currStatus, invalidStatus, response.Code))
		}
	}

	// try various invalid images status updates
	imagesStatusTests := map[model.ImagesStatus][]model.ImagesStatus{
		model.ImagesStatusDefault: {
			model.ImagesStatusToLoad,
			model.ImagesStatusLoading,
			model.ImagesStatusLoadComplete,
			model.ImagesStatusLoadError,
			model.ImagesStatusError,
		},
		model.ImagesStatusLoading: {
			model.ImagesStatusDefault,
			model.ImagesStatusToLoad,
			model.ImagesStatusLoadComplete,
			model.ImagesStatusLoadError,
			model.ImagesStatusError,
		},
		model.ImagesStatusError: {
			model.ImagesStatusDefault,
			model.ImagesStatusToLoad,
			model.ImagesStatusLoading,
			model.ImagesStatusLoadComplete,
			model.ImagesStatusLoadError,
		},
	}
	for currStatus, invalidStatuses := range imagesStatusTests {
		post.PostStatus = model.PostStatusDraft
		post.RecordsStatus = model.RecordsStatusDefault
		post.ImagesStatus = currStatus
		post, err = postPersister.UpdatePost(ctx, post.ID, *post)
		assert.NoError(t, err, fmt.Sprintf("updating images status %s", currStatus))

		for _, invalidStatus := range invalidStatuses {
			post.ImagesStatus = invalidStatus
			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			err := enc.Encode(post)
			if err != nil {
				t.Errorf("Error encoding PostIn: %v", err)
			}
			request, _ := http.NewRequest("PUT", fmt.Sprintf("/societies/1/posts/%d", post.ID), buf)
			request.Header.Add("Content-Type", contentType)
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)
			assert.Equal(t, http.StatusBadRequest, response.Code,
				fmt.Sprintf("updating images status from %s to %s expected bad request; response is: %d",
					currStatus, invalidStatus, response.Code))
		}
	}
}

func TestDeletePost(t *testing.T) {
	am := &api.ApiMock{}
	app := NewApp().API(am)
	app.authDisabled = true
	r := app.NewRouter()

	am.Result = nil
	am.Errors = nil

	request, _ := http.NewRequest("DELETE", "/societies/1/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code, "Response: %s", string(response.Body.Bytes()))
}

func makePostIn(t *testing.T, collectionID uint32) (model.PostIn, *bytes.Buffer) {
	in := model.PostIn{
		PostBody: model.PostBody{
			Name:       "First",
			PostStatus: model.PostStatusDraft,
		},
		Collection: collectionID,
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(in)
	if err != nil {
		t.Errorf("Error encoding PostIn: %v", err)
	}
	return in, buf
}
