package api

import (
	"context"
	"log"

	"github.com/ourrootsorg/cms-server/model"
)

// CollectionResult is a paged Collection result
type CollectionResult struct {
	Collections []model.Collection `json:"collections"`
	NextPage    string             `json:"next_page"`
}

// GetCollections holds the business logic around getting all Collections
func (api API) GetCollections(ctx context.Context /* filter/search criteria */) (*CollectionResult, error) {
	// TODO: handle search criteria and paged results
	cols, err := api.collectionPersister.SelectCollections(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	return &CollectionResult{Collections: cols}, nil
}

// GetCollectionsByID holds the business logic around getting many Collections
func (api API) GetCollectionsByID(ctx context.Context, ids []uint32) ([]model.Collection, error) {
	colls, err := api.collectionPersister.SelectCollectionsByID(ctx, ids)
	if err != nil {
		return nil, NewError(err)
	}
	return colls, nil
}

// GetCollection holds the business logic around getting a Collection
func (api API) GetCollection(ctx context.Context, id uint32) (*model.Collection, error) {
	collection, err := api.collectionPersister.SelectOneCollection(ctx, id)
	if err != nil {
		return nil, NewError(err)
	}
	return collection, nil
}

// AddCollection holds the business logic around adding a Collection
func (api API) AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, error) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid collection %v", err)
		return nil, NewError(err)
	}
	collection, e := api.collectionPersister.InsertCollection(ctx, in)
	if e != nil {
		return nil, NewError(e)
	}
	return collection, nil
}

// UpdateCollection holds the business logic around updating a Collection
func (api API) UpdateCollection(ctx context.Context, id uint32, in model.Collection) (*model.Collection, error) {
	err := api.validate.Struct(in)
	log.Printf("[DEBUG] Collection=%v err=%v\n", in, err)
	if err != nil {
		return nil, NewError(err)
	}
	collection, e := api.collectionPersister.UpdateCollection(ctx, id, in)
	if e != nil {
		return nil, NewError(e)
	}
	return collection, nil
}

// DeleteCollection holds the business logic around deleting a Collection
func (api API) DeleteCollection(ctx context.Context, id uint32) error {
	err := api.collectionPersister.DeleteCollection(ctx, id)
	if err != nil {
		return NewError(err)
	}
	return nil
}
