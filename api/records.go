package api

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
)

// RecordResult is a paged Record result
type RecordResult struct {
	Records  []model.Record `json:"records"`
	NextPage string         `json:"next_page"`
}

// GetRecords holds the business logic around getting all Records for a post
func (api API) GetRecordsForPost(ctx context.Context, postID uint32) (*RecordResult, *model.Errors) {
	// TODO: handle search criteria and paged results
	records, err := api.recordPersister.SelectRecordsForPost(ctx, postID)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &RecordResult{Records: records}, nil
}

// GetRecordsByID holds the business logic around getting many Records
func (api API) GetRecordsByID(ctx context.Context, ids []uint32) ([]model.Record, *model.Errors) {
	records, err := api.recordPersister.SelectRecordsByID(ctx, ids)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return records, nil
}

// GetRecord holds the business logic around getting a Record
func (api API) GetRecord(ctx context.Context, id uint32) (*model.Record, *model.Errors) {
	record, err := api.recordPersister.SelectOneRecord(ctx, id)
	if err == persist.ErrNoRows {
		return nil, model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, strconv.Itoa(int(id))))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &record, nil
}

// AddRecord holds the business logic around adding a Record
func (api API) AddRecord(ctx context.Context, in model.RecordIn) (*model.Record, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid record %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	// insert
	record, err := api.recordPersister.InsertRecord(ctx, in)
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid post reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, strconv.Itoa(int(in.Post)), "post"))
	} else if err != nil {
		log.Printf("[ERROR] Internal server error: %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &record, nil
}

// UpdateRecord holds the business logic around updating a Record
func (api API) UpdateRecord(ctx context.Context, id uint32, in model.Record) (*model.Record, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	record, err := api.recordPersister.UpdateRecord(ctx, id, in)
	if er, ok := err.(model.Error); ok {
		if er.Code == model.ErrConcurrentUpdate {
			return nil, model.NewErrors(http.StatusConflict, er)
		} else if er.Code == model.ErrNotFound {
			// Not allowed to add a Record with PUT
			return nil, model.NewErrors(http.StatusNotFound, er)
		}
	}
	if err == persist.ErrForeignKeyViolation {
		log.Printf("[ERROR] Invalid post reference: %v", err)
		return nil, model.NewErrors(http.StatusBadRequest, model.NewError(model.ErrBadReference, strconv.Itoa(int(in.Post)), "post"))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &record, nil
}

// DeleteRecord holds the business logic around deleting a Record
func (api API) DeleteRecord(ctx context.Context, id uint32) *model.Errors {
	err := api.recordPersister.DeleteRecord(ctx, id)
	if err != nil {
		return model.NewErrors(http.StatusInternalServerError, err)
	}
	return nil
}

// DeleteRecord holds the business logic around deleting a Record
func (api API) DeleteRecordsForPost(ctx context.Context, postID uint32) *model.Errors {
	err := api.recordPersister.DeleteRecordsForPost(ctx, postID)
	if err != nil {
		return model.NewErrors(http.StatusInternalServerError, err)
	}
	return nil
}
