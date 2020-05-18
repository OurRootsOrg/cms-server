package api

import (
	"context"
	"log"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
)

// PostResult is a paged Post result
type PostResult struct {
	Posts    []model.Post `json:"posts"`
	NextPage string       `json:"next_page"`
}

// GetPosts holds the business logic around getting many Posts
func (api API) GetPosts(ctx context.Context /* filter/search criteria */) (*PostResult, *model.Errors) {
	// TODO: handle search criteria and paged results
	posts, err := api.postPersister.SelectPosts(ctx)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &PostResult{Posts: posts}, nil
}

// GetPost holds the business logic around getting a Post
func (api API) GetPost(ctx context.Context, id string) (*model.Post, *model.Errors) {
	post, err := api.postPersister.SelectOnePost(ctx, id)
	if err == persist.ErrNoRows {
		return nil, model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, id))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &post, nil
}

// AddPost holds the business logic around adding a Post
func (api API) AddPost(ctx context.Context, in model.PostIn) (*model.Post, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid post %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	post, err := api.postPersister.InsertPost(ctx, in)
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid collection reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, in.Collection, "collection"))
	} else if err != nil {
		log.Printf("[ERROR] Internal server error: %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &post, nil
}

// UpdatePost holds the business logic around updating a Post
func (api API) UpdatePost(ctx context.Context, id string, in model.Post) (*model.Post, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	post, err := api.postPersister.UpdatePost(ctx, id, in)
	if er, ok := err.(model.Error); ok {
		if er.Code == model.ErrConcurrentUpdate {
			return nil, model.NewErrors(http.StatusConflict, er)
		} else if er.Code == model.ErrNotFound {
			// Not allowed to add a Post with PUT
			return nil, model.NewErrors(http.StatusNotFound, er)
		}
	}
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid collection reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, in.Collection, "collection"))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &post, nil
}

// DeletePost holds the business logic around deleting a Post
func (api API) DeletePost(ctx context.Context, id string) *model.Errors {
	err := api.postPersister.DeletePost(ctx, id)
	if err != nil {
		return model.NewErrors(http.StatusInternalServerError, err)
	}
	return nil
}