package api

import (
	"context"
	"log"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
)

// CollectionResult is a paged Collection result
type CollectionResult struct {
	Collections []model.Collection `json:"collections"`
	NextPage    string             `json:"next_page"`
}

// GetCollections holds the business logic around getting all Collections
func (api API) GetCollections(ctx context.Context /* filter/search criteria */) (*CollectionResult, *model.Errors) {
	// TODO: handle search criteria and paged results
	cols, err := api.collectionPersister.SelectCollections(ctx)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &CollectionResult{Collections: cols}, nil
}

// GetManyCollections holds the business logic around getting many Collections
func (api API) GetManyCollections(ctx context.Context, ids []string) ([]model.Collection, *model.Errors) {
	colls, err := api.collectionPersister.SelectManyCollections(ctx, ids)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return colls, nil
}

// GetCollection holds the business logic around getting a Collection
func (api API) GetCollection(ctx context.Context, id string) (*model.Collection, *model.Errors) {
	collection, err := api.collectionPersister.SelectOneCollection(ctx, id)
	if err == persist.ErrNoRows {
		return nil, model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, id))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// AddCollection holds the business logic around adding a Collection
func (api API) AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid collection %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	collection, err := api.collectionPersister.InsertCollection(ctx, in)
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid category reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, in.Category, "category"))
	} else if err != nil {
		log.Printf("[ERROR] Internal server error: %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// UpdateCollection holds the business logic around updating a Collection
func (api API) UpdateCollection(ctx context.Context, id string, in model.Collection) (*model.Collection, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	collection, err := api.collectionPersister.UpdateCollection(ctx, id, in)
	if er, ok := err.(model.Error); ok {
		if er.Code == model.ErrConcurrentUpdate {
			return nil, model.NewErrors(http.StatusConflict, er)
		} else if er.Code == model.ErrNotFound {
			// Not allowed to add a Collection with PUT
			return nil, model.NewErrors(http.StatusNotFound, er)
		}
	}
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid category reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, in.Category, "category"))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// DeleteCollection holds the business logic around deleting a Collection
func (api API) DeleteCollection(ctx context.Context, id string) *model.Errors {
	err := api.collectionPersister.DeleteCollection(ctx, id)
	if err != nil {
		return model.NewErrors(http.StatusInternalServerError, err)
	}
	return nil
}
