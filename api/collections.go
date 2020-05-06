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

// GetCollections holds the business logic around getting many Collections
func (api API) GetCollections(ctx context.Context /* filter/search criteria */) (*CollectionResult, *Errors) {
	// TODO: handle search criteria and paged results
	cols, err := api.collectionPersister.SelectCollections(ctx)
	if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &CollectionResult{Collections: cols}, nil
}

// GetCollection holds the business logic around getting a Collection
func (api API) GetCollection(ctx context.Context, id string) (*model.Collection, *Errors) {
	collection, err := api.collectionPersister.SelectOneCollection(ctx, id)
	if err == persist.ErrNoRows {
		return nil, NewErrors(http.StatusNotFound, NewError(ErrNotFound, id))
	} else if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// AddCollection holds the business logic around adding a Collection
func (api API) AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, *Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewErrors(http.StatusBadRequest, err)
	}
	collection, err := api.collectionPersister.InsertCollection(ctx, in)
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid category reference: %v", err)
		return nil, NewErrors(http.StatusBadRequest, NewError(ErrBadReference, in.Category.ID, in.Category.Type))
	} else if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// UpdateCollection holds the business logic around updating a Collection
func (api API) UpdateCollection(ctx context.Context, id string, in model.CollectionIn) (*model.Collection, *Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewErrors(http.StatusBadRequest, err)
	}
	collection, err := api.collectionPersister.UpdateCollection(ctx, id, in)
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid category reference: %v", err)
		return nil, NewErrors(http.StatusBadRequest, NewError(ErrBadReference, in.Category.ID, in.Category.Type))
	} else if err == persist.ErrNoRows {
		// Not allowed to add a Collection with PUT
		return nil, NewErrors(http.StatusNotFound, NewError(ErrNotFound, id))
	} else if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// DeleteCollection holds the business logic around deleting a Collection
func (api API) DeleteCollection(ctx context.Context, id string) *Errors {
	err := api.collectionPersister.DeleteCollection(ctx, id)
	if err != nil {
		return NewErrors(http.StatusInternalServerError, err)
	}
	return nil
}
