package api

import (
	"context"
	"log"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// RecordResult is a paged Record result
type RecordResult struct {
	Records  []model.Record `json:"records"`
	NextPage string         `json:"next_page"`
}

// GetRecordsForPost holds the business logic around getting all Records for a post
func (api API) GetRecordsForPost(ctx context.Context, postID uint32) (*RecordResult, error) {
	// TODO: handle search criteria and paged results
	records, err := api.recordPersister.SelectRecordsForPost(ctx, postID)
	if err != nil {
		return nil, NewErrors(0, err)
	}
	return &RecordResult{Records: records}, nil
}

// GetRecordsByID holds the business logic around getting many Records
func (api API) GetRecordsByID(ctx context.Context, ids []uint32) ([]model.Record, error) {
	records, err := api.recordPersister.SelectRecordsByID(ctx, ids)
	if err != nil {
		return nil, NewErrors(0, err)
	}
	return records, nil
}

// GetRecord holds the business logic around getting a Record
func (api API) GetRecord(ctx context.Context, id uint32) (*model.Record, error) {
	record, err := api.recordPersister.SelectOneRecord(ctx, id)
	if err != nil {
		return nil, NewErrors(0, err)
	}
	return record, nil
}

// AddRecord holds the business logic around adding a Record
func (api API) AddRecord(ctx context.Context, in model.RecordIn) (*model.Record, error) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid record %v", err)
		return nil, NewErrors(http.StatusBadRequest, err)
	}
	// insert
	record, e := api.recordPersister.InsertRecord(ctx, in)
	if e != nil {
		return nil, NewErrors(0, e)
	}
	return record, nil
}

// UpdateRecord holds the business logic around updating a Record
func (api API) UpdateRecord(ctx context.Context, id uint32, in model.Record) (*model.Record, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewErrors(http.StatusBadRequest, err)
	}
	record, e := api.recordPersister.UpdateRecord(ctx, id, in)
	if e != nil {
		return nil, NewErrors(0, e)
	}
	return record, nil
}

// DeleteRecord holds the business logic around deleting a Record
func (api API) DeleteRecord(ctx context.Context, id uint32) error {
	err := api.recordPersister.DeleteRecord(ctx, id)
	if err != nil {
		return NewErrors(0, err)
	}
	return nil
}

// DeleteRecordsForPost holds the business logic around deleting the Records for a Post
func (api API) DeleteRecordsForPost(ctx context.Context, postID uint32) error {
	err := api.recordPersister.DeleteRecordsForPost(ctx, postID)
	if err != nil {
		return NewErrors(0, err)
	}
	return nil
}
